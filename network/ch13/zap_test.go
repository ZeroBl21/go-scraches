package ch13

import (
	"bytes"
	"fmt"
	"os"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var encoderConfig = zapcore.EncoderConfig{
	MessageKey: "msg",
	NameKey:    "name",

	LevelKey:    "level",
	EncodeLevel: zapcore.LowercaseLevelEncoder,

	CallerKey:    "caller",
	EncodeCaller: zapcore.ShortCallerEncoder,

	// TimeKey: "time",
	// EncodeTime: zapcore.ISO8601TimeEncoder,
}

func Example_zapJSON() {
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.Lock(os.Stdout),
			zapcore.DebugLevel,
		),
		zap.AddCaller(),
		zap.Fields(
			zap.String("version", runtime.Version()),
		),
	)
	defer zl.Sync()

	example := zl.Named("example")
	example.Debug("test debug message")
	example.Info("test info message")

	// Output:
	// {"level":"debug","name":"example","caller":"ch13/zap_test.go:42","msg":"test debug message","version":"go1.23.1"}
	// {"level":"info","name":"example","caller":"ch13/zap_test.go:43","msg":"test info message","version":"go1.23.1"}
}

func Example_zapConsole() {
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.Lock(os.Stdout),
			zapcore.InfoLevel,
		),
	)
	defer zl.Sync()

	console := zl.Named("[console]")
	console.Info("this is logged by the logger")
	console.Debug("this is bellow the logger's threshold and won't log")
	console.Error("this is also logged by the logger")

	// Output:
	// info	[console]	this is logged by the logger
	// error	[console]	this is also logged by the logger
}

func Example_zapInfoFileDebugConsole() {
	logFile := new(bytes.Buffer)

	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.Lock(zapcore.AddSync(logFile)),
			zapcore.InfoLevel,
		),
	)
	defer zl.Sync()

	zl.Debug("this is below the logger's threshold and won't log")
	zl.Error("this is logged by the logger")

	zl = zl.WithOptions(
		zap.WrapCore(
			func(c zapcore.Core) zapcore.Core {
				ucEncoderConfig := encoderConfig
				ucEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

				return zapcore.NewTee(c, zapcore.NewCore(
					zapcore.NewConsoleEncoder(ucEncoderConfig),
					zapcore.Lock(os.Stdout),
					zapcore.DebugLevel,
				))
			},
		),
	)

	fmt.Println("standard output:")
	zl.Debug("this is only logged as console encoding")
	zl.Info("this is logged as console encoding and JSON")

	fmt.Print("\nlog file contents:\n", logFile.String())

	// Output:
	// standard output:
	// DEBUG	this is only logged as console encoding
	// INFO	this is logged as console encoding and JSON
	//
	// log file contents:
	// {"level":"error","msg":"this is logged by the logger"}
	// {"level":"info","msg":"this is logged as console encoding and JSON"}
}
