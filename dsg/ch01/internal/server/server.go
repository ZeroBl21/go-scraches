package server

import (
	"context"
	"errors"
	"time"

	api "github.com/ZeroBl21/dsg/ch01/proglog/api/v1"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const (
	objectWildCart = "*"
	produceAction  = "produce"
	consumeAction  = "consume"
)

type Config struct {
	CommitLog  CommitLog
	Authorizer Authorizer
}

type CommitLog interface {
	Append(*api.Record) (uint64, error)
	Read(uint64) (*api.Record, error)
}

type Authorizer interface {
	Authorize(subject, object, action string) error
}

var _ api.LogServer = (*grpcServer)(nil)

type grpcServer struct {
	api.UnimplementedLogServer
	*Config
}

func NewGRPCServer(config *Config, grpcOpts ...grpc.ServerOption) (*grpc.Server, error) {
	logger := zap.L().Named("server")
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Int64("grpc.time_ns", duration.Nanoseconds())
		}),
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		return nil, err
	}

	grpcOpts = append(
		grpcOpts,
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(logger, zapOpts...),
			grpc_auth.StreamServerInterceptor(authenticate)),
		),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
			grpc_auth.UnaryServerInterceptor(authenticate)),
		),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)

	gSrv := grpc.NewServer(grpcOpts...)
	srv, err := newGRPCServer(config)
	if err != nil {
		return nil, err
	}

	api.RegisterLogServer(gSrv, srv)

	return gSrv, nil
}

func newGRPCServer(config *Config) (*grpcServer, error) {
	srv := &grpcServer{Config: config}

	return srv, nil
}

func (s *grpcServer) Produce(
	ctx context.Context,
	req *api.ProduceRequest,
) (*api.ProduceResponse, error) {
	err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildCart,
		produceAction,
	)
	if err != nil {
		return nil, err
	}

	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}

	return &api.ProduceResponse{Offset: offset}, nil
}

func (s *grpcServer) Consume(
	ctx context.Context,
	req *api.ConsumeRequest,
) (*api.ConsumeResponse, error) {
	err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildCart,
		consumeAction,
	)
	if err != nil {
		return nil, err
	}

	record, err := s.CommitLog.Read(req.Offset)
	if err != nil {
		return nil, err
	}

	return &api.ConsumeResponse{Record: record}, nil
}

func (s *grpcServer) ProduceStream(
	stream api.Log_ProduceStreamServer,
) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		res, err := s.Produce(stream.Context(), req)
		if err != nil {
			return err
		}

		if err := stream.Send(res); err != nil {
			return err
		}
	}
}

func (s *grpcServer) ConsumeStream(
	req *api.ConsumeRequest,
	stream api.Log_ConsumeStreamServer,
) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			res, err := s.Consume(stream.Context(), req)
			if err != nil {
				if errors.Is(err, api.ErrOffsetOutOfRange{}) {
					continue
				}
				return err
			}

			if err := stream.Send(res); err != nil {
				return err
			}
			req.Offset++
		}
	}
}

func authenticate(ctx context.Context) (context.Context, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(
			codes.Unknown,
			"couldn't find peer info",
		).Err()
	}

	if peer.AuthInfo == nil {
		return context.WithValue(ctx, subjectContextKey{}, ""), nil
	}

	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	ctx = context.WithValue(ctx, subjectContextKey{}, subject)

	return ctx, nil
}

type subjectContextKey struct{}

func subject(ctx context.Context) string {
	return ctx.Value(subjectContextKey{}).(string)
}
