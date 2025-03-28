package server

import (
	"context"
	"net"
	"testing"

	api "github.com/ZeroBl21/dsg/ch01/proglog/api/v1"
	"github.com/ZeroBl21/dsg/ch01/proglog/internal/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		client api.LogClient,
		config *Config,
	){
		"produce/consume a message to/from the log succeeds": testProduceConsume,
		"produce/consume stream succeeds":                    testProduceConsumeStream,
		"consume past log boundary fails":                    testConsumePastBoundary,
	} {
		t.Run(scenario, func(t *testing.T) {
			client, config, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, client, config)
		})
	}
}

func setupTest(t *testing.T, fn func(*Config)) (
	client api.LogClient,
	cfg *Config,
	teardown func(),
) {
	t.Helper()

	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	clientOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	grpcClient, err := grpc.NewClient(l.Addr().String(), clientOptions...)
	require.NoError(t, err)

	dir := t.TempDir()

	cLog, err := log.NewLog(dir, log.Config{})
	require.NoError(t, err)

	cfg = &Config{
		CommitLog: cLog,
	}
	if fn != nil {
		fn(cfg)
	}

	server, err := NewGRPCServer(cfg)
	require.NoError(t, err)

	go func() {
		server.Serve(l)
	}()

	client = api.NewLogClient(grpcClient)

	return client, cfg, func() {
		server.Stop()
		grpcClient.Close()
		l.Close()
		cLog.Close()
	}
}

func testProduceConsume(
	t *testing.T,
	client api.LogClient,
	cfg *Config,
) {
	ctx := context.Background()

	want := &api.Record{Value: []byte("hello world")}

	produce, err := client.Produce(ctx, &api.ProduceRequest{Record: want})
	require.NoError(t, err)

	consume, err := client.Consume(ctx, &api.ConsumeRequest{Offset: produce.Offset})
	require.NoError(t, err)
	require.Equal(t, want.Value, consume.Record.Value)
	require.Equal(t, want.Offset, consume.Record.Offset)
}

func testConsumePastBoundary(
	t *testing.T,
	client api.LogClient,
	cfg *Config,
) {
	ctx := context.Background()

	produce, err := client.Produce(
		ctx,
		&api.ProduceRequest{Record: &api.Record{Value: []byte("hello world")}},
	)
	require.NoError(t, err)

	consume, err := client.Consume(
		ctx,
		&api.ConsumeRequest{Offset: produce.Offset + 1},
	)
	if consume != nil {
		t.Fatal("consume not nil")
	}

	got := status.Code(err)
	want := status.Code(api.ErrOffsetOutOfRange{}.GRPCStatus().Err())

	if got != want {
		t.Fatalf("got err: %v, want: %v instead", got, want)
	}
}

func testProduceConsumeStream(
	t *testing.T,
	client api.LogClient,
	cfg *Config,
) {
	ctx := context.Background()

	records := []*api.Record{
		{Value: []byte("first message")},
		{Value: []byte("second message")},
	}

	// Produce
	produceStream, err := client.ProduceStream(ctx)
	require.NoError(t, err)

	for offset, record := range records {
		err = produceStream.Send(&api.ProduceRequest{
			Record: record,
		})
		require.NoError(t, err)

		res, err := produceStream.Recv()
		require.NoError(t, err)

		if res.Offset != uint64(offset) {
			t.Fatalf("got offset: %d, want: %d",
				res.Offset,
				offset,
			)
		}
	}

	// Consume
	consumeStream, err := client.ConsumeStream(ctx, &api.ConsumeRequest{Offset: 0})
	require.NoError(t, err)

	for i, record := range records {
		res, err := consumeStream.Recv()
		require.NoError(t, err)
		require.Equal(t, res.Record, &api.Record{
			Value:  record.Value,
			Offset: uint64(i),
		})
	}
}
