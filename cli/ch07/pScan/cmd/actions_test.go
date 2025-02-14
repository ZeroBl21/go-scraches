package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ZeroBl21/cli/ch07/pScan/scan"
)

func TestHostActions(t *testing.T) {
	hosts := []string{
		"host1",
		"host2",
		"host3",
	}

	testCases := []struct {
		name           string
		args           []string
		expectedOut    string
		initList       bool
		actionFunction func(io.Writer, string, []string) error
	}{
		{
			name:           "AddAction",
			args:           hosts,
			expectedOut:    "Added host: host1\nAdded host: host2\nAdded host: host3\n",
			initList:       false,
			actionFunction: addAction,
		},
		{
			name:           "ListAction",
			expectedOut:    "host1\nhost2\nhost3\n",
			initList:       true,
			actionFunction: listAction,
		},
		{
			name:           "DeleteAction",
			args:           []string{"host1", "host2"},
			expectedOut:    "Deleted host: host1\nDeleted host: host2\n",
			initList:       true,
			actionFunction: deleteAction,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tf, cleanup := setup(t, hosts, tc.initList)
			defer cleanup()

			var out bytes.Buffer

			if err := tc.actionFunction(&out, tf, tc.args); err != nil {
				t.Fatalf("Expected no error, got %q\n", err)
			}

			if out.String() != tc.expectedOut {
				t.Errorf("Expected output %q, got %q\n", tc.expectedOut, out.String())
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	hosts := []string{
		"host1",
		"host2",
		"host3",
	}

	tf, cleanup := setup(t, hosts, false)
	defer cleanup()

	delHost := "host2"

	hostsEnd := []string{
		"host1",
		"host3",
	}

	var out bytes.Buffer
	var expectedOut strings.Builder

	// add -> list -> delete -> list -> scan

	for _, v := range hosts {
		fmt.Fprintf(&expectedOut, "Added host: %s\n", v)
	}

	expectedOut.WriteString(strings.Join(hosts, "\n") + "\n")
	fmt.Fprintf(&expectedOut, "Deleted host: %s\n", delHost)
	expectedOut.WriteString(strings.Join(hostsEnd, "\n") + "\n")

	for _, v := range hostsEnd {
		fmt.Fprintf(&expectedOut, "%s: Host not found\n\n", v)
	}

	if err := addAction(&out, tf, hosts); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	if err := listAction(&out, tf, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	if err := deleteAction(&out, tf, []string{delHost}); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	if err := listAction(&out, tf, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	if err := scanAction(&out, tf, nil, time.Duration(1000)); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	if out.String() != expectedOut.String() {
		t.Errorf("Expected output %q, got %q\n", expectedOut.String(), out.String())
	}
}

func setup(t *testing.T, hosts []string, initList bool) (string, func()) {
	tf, err := os.CreateTemp("", "pScan")
	if err != nil {
		t.Fatal(err)
	}
	tf.Close()

	if initList {
		hl := &scan.HostsList{}

		for _, h := range hosts {
			hl.Add(h)
		}

		if err := hl.Save(tf.Name()); err != nil {
			t.Fatal(err)
		}
	}

	return tf.Name(), func() {
		os.Remove(tf.Name())
	}
}

func TestScanAction(t *testing.T) {
	hosts := []string{
		"localhost",
		"unknownhostoutthere",
	}

	tf, cleanup := setup(t, hosts, true)
	defer cleanup()

	ports := []int{}
	timeout := time.Duration(1000)

	for i := 0; i < len(hosts); i++ {
		ln, err := net.Listen("tcp", net.JoinHostPort("localhost", "0"))
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()

		_, portStr, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			t.Fatal(err)
		}

		ports = append(ports, port)

		// If host is "unknownhostoutthere" close listener
		if i == 1 {
			ln.Close()
		}
	}

	var expectedOut strings.Builder

	expectedOut.WriteString("localhost:\n")
	expectedOut.WriteString(fmt.Sprintf("\t%d: open\n", ports[0]))
	expectedOut.WriteString(fmt.Sprintf("\t%d: closed\n", ports[1]))
	expectedOut.WriteString("\n")
	expectedOut.WriteString(fmt.Sprintf("%s: Host not found\n\n", hosts[1]))

	var out bytes.Buffer

	if err := scanAction(&out, tf, ports, timeout); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	if out.String() != expectedOut.String() {
		t.Errorf("Expected output %q, got %q\n", &expectedOut, out.String())
	}
}
