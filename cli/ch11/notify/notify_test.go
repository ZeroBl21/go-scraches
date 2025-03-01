//go:build !integration
// +build !integration

package notify

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		s Severity
	}{
		{SeverityLow},
		{SeverityNormal},
		{SeverityUrgent},
	}

	for _, tt := range testCases {
		name := tt.s.String()

		expMsg := "Message"
		expTitle := "Title"

		t.Run(name, func(t *testing.T) {
			n := New(expTitle, expMsg, tt.s)

			if n.message != expMsg {
				t.Errorf("Expected %q, got %q instead\n", expMsg, n.message)
			}

			if n.title != expTitle {
				t.Errorf("Expected %q, got %q instead\n", expTitle, n.title)
			}

			if n.severity != tt.s {
				t.Log(tt.s)
				t.Log(n.severity)
				t.Errorf("Expected %q, got %q instead\n", tt.s, n.severity)
			}
		})
	}
}

func TestSeverityString(t *testing.T) {
	testCases := []struct {
		s   Severity
		exp string
		os  string
	}{
		// Linux
		{SeverityLow, "low", "linux"},
		{SeverityNormal, "normal", "linux"},
		{SeverityUrgent, "critical", "linux"},

		// Mac
		{SeverityLow, "Low", "darwin"},
		{SeverityNormal, "Normal", "darwin"},
		{SeverityUrgent, "Critical", "darwin"},

		// Windows
		{SeverityLow, "Info", "windows"},
		{SeverityNormal, "Warning", "windows"},
		{SeverityUrgent, "Error", "windows"},
	}

	for _, tt := range testCases {
		name := fmt.Sprintf("%s%d", tt.os, tt.s)

		t.Run(name, func(t *testing.T) {
			if runtime.GOOS != tt.os {
				t.Skip("Skipped: not OS", runtime.GOOS)
			}

			sev := tt.s.String()
			if sev != tt.exp {
				t.Errorf("Expected %q, got %q instead\n", tt.exp, sev)
			}
		})
	}
}

func TestSend(t *testing.T) {
	n := New("test title", "test msg", SeverityNormal)

	command = mockCmd

	if err := n.Send(); err != nil {
		t.Error(err)
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	cmdName := ""

	switch runtime.GOOS {
	case "linux":
		cmdName = "notify-send"
	case "darwin":
		cmdName = "terminal-notifier"
	case "windows":
		cmdName = "powershell"
	}

	if strings.Contains(os.Args[2], cmdName) {
		os.Exit(0)
	}

	os.Exit(1)
}

func mockCmd(exe string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess"}
	cs = append(cs, exe)
	cs = append(cs, args...)

	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}

	return cmd
}
