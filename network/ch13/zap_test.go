package ch13

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/fsnotify.v1"
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
	// {"level":"debug","name":"example","caller":"ch13/zap_test.go:46","msg":"test debug message","version":"go1.23.1"}
	// {"level":"info","name":"example","caller":"ch13/zap_test.go:47","msg":"test info message","version":"go1.23.1"}
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

func Example_zapSampling() {
	zl := zap.New(
		zapcore.NewSamplerWithOptions(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				zapcore.Lock(os.Stdout),
				zapcore.DebugLevel,
			),
			time.Second, 1, 3,
		),
	)
	defer zl.Sync()

	for i := 0; i < 10; i++ {
		if i == 5 {
			time.Sleep(time.Second)
		}

		zl.Debug(fmt.Sprintf("%d", i))
		zl.Debug("debug message")
	}

	// Output:
	// {"level":"debug","msg":"0"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"1"}
	// {"level":"debug","msg":"2"}
	// {"level":"debug","msg":"3"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"4"}
	// {"level":"debug","msg":"5"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"6"}
	// {"level":"debug","msg":"7"}
	// {"level":"debug","msg":"8"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"9"}
}

func Example_zapDynamicDebugging() {
	tempDir := os.TempDir()
	defer os.RemoveAll(tempDir)

	debugLevelFile := filepath.Join(tempDir, "level.debug")
	atomicLevel := zap.NewAtomicLevel()

	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.Lock(os.Stdout),
			atomicLevel,
		),
	)
	defer zl.Sync()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	if err := watcher.Add(tempDir); err != nil {
		log.Fatal(err)
	}

	ready := make(chan struct{})

	go func() {
		defer close(ready)

		originalLevel := atomicLevel.Level()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Name == debugLevelFile {
					switch {
					case event.Op&fsnotify.Create == fsnotify.Create:
						atomicLevel.SetLevel(zapcore.DebugLevel)
						ready <- struct{}{}

					case event.Op&fsnotify.Remove == fsnotify.Remove:
						atomicLevel.SetLevel(originalLevel)
						ready <- struct{}{}
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				zl.Error(err.Error())
			}
		}
	}()

	zl.Debug("this is below the logger's threshold")

	df, err := os.Create(debugLevelFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := df.Close(); err != nil {
		log.Fatal(err)
	}

	<-ready

	zl.Debug("this is now at the logger's threshold")

	if err := os.Remove(debugLevelFile); err != nil {
		log.Fatal(err)
	}
	<-ready

	zl.Debug("this is bellow the logger's threshold again")
	zl.Info("this is at the logger's current threshold")

	// Output:
	// {"level":"debug","msg":"this is now at the logger's threshold"}
	// {"level":"info","msg":"this is at the logger's current threshold"}
}
