package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	proj := flag.String("p", "", "Project directory")
	branch := flag.String("b", "main", "Git branch to push the code")

	flag.Parse()

	if err := run(*proj, *branch, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

type executer interface {
	execute() (string, error)
}

func run(proj, branch string, out io.Writer) error {
	if proj == "" {
		return fmt.Errorf("Project directory is required: %w", ErrValidation)
	}

	if branch == "" {
		return fmt.Errorf("Git branch is required: %w", ErrValidation)
	}

	pipeline := make([]executer, 4)

	pipeline[0] = newStep(
		"go build",
		"go",
		"Go Build: SUCCESS",
		proj,
		[]string{"build", ".", "errors"},
	)

	pipeline[1] = newStep(
		"go test",
		"go",
		"Go Test: SUCCESS",
		proj,
		[]string{"test", "-v"},
	)

	pipeline[2] = newExceptionStep(
		"go fmt",
		"gofmt",
		"Gofmt: SUCCESS",
		proj,
		[]string{"-l", "."},
	)

	pipeline[3] = newTimeoutStep(
		"git push",
		"git",
		"Git Push: SUCCESS",
		proj,
		[]string{"push", "origin", branch},
		10*time.Second,
	)

	sig := make(chan os.Signal, 1)
	errCh := make(chan error)
	done := make(chan struct{})

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for _, s := range pipeline {
			msg, err := s.execute()
			if err != nil {
				errCh <- err
				return
			}

			_, err = fmt.Fprintln(out, msg)
			if err != nil {
				errCh <- err
				return
			}
		}

		close(done)
	}()

	for {
		select {
		case rec := <-sig:
			signal.Stop(sig)
			return fmt.Errorf("%s: Exiting: %w", rec, ErrSignal)

		case err := <-errCh:
			return err

		case <-done:
			return nil
		}
	}
}
