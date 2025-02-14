package scan_test

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/ZeroBl21/cli/ch07/pScan/scan"
)

func TestStateString(t *testing.T) {
	ps := scan.PortState{}

	if ps.Open.String() != scan.PORT_CLOSED {
		t.Errorf("Expected %q, got %q instead\n", "closed", ps.Open.String())
	}

	ps.Open = true

	if ps.Open.String() != scan.PORT_OPEN {
		t.Errorf("Expected %q, got %q instead\n", "open", ps.Open.String())
	}
}

func TestRunHostFound(t *testing.T) {
	testCases := []struct {
		name          string
		expectedState string
	}{
		{"OpenPort", scan.PORT_OPEN},
		{"ClosedPort", scan.PORT_CLOSED},
	}

	host := "localhost"

	hl := &scan.HostsList{}
	hl.Add(host)

	ports := []int{}
	timeout := time.Duration(1000)

	for _, tc := range testCases {
		ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
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

		if tc.name == "ClosedPort" {
			ln.Close()
		}
	}

	res := scan.Run(hl, ports, timeout)

	if len(res) != 1 {
		t.Fatalf("Expected 1 results, got %d instead\n", len(res))
	}

	if res[0].Host != host {
		t.Errorf("Expected host %q, got %q instead\n", host, res[0].Host)
	}

	if res[0].NotFound {
		t.Errorf("Expected host %q to be found\n", host)
	}

	if len(res[0].PortStates) != 2 {
		t.Fatalf("Expected 2 port states, got %d instead\n", len(res[0].PortStates))
	}

	for idx, tc := range testCases {
		if res[0].PortStates[idx].Port != ports[idx] {
			t.Errorf("Expected port %d, got %d instead\n",
				ports[idx], res[0].PortStates[idx].Port)
		}

		if res[0].PortStates[idx].Open.String() != tc.expectedState {
			t.Errorf("Expected port %d to be %s", ports[idx], tc.expectedState)
		}
	}
}

func TestRunHostNotFound(t *testing.T) {
	host := "389.389.389.389"
	timeout := time.Duration(1000)

	hl := &scan.HostsList{}
	hl.Add(host)

	res := scan.Run(hl, []int{}, timeout)

	if len(res) != 1 {
		t.Fatalf("Expected 1 results, got %d instead\n", len(res))
	}

	if res[0].Host != host {
		t.Errorf("Expected host %q, got %q instead\n", host, res[0].Host)
	}

	if !res[0].NotFound {
		t.Errorf("Expected host %q NOT to be found\n", host)
	}

	if len(res[0].PortStates) != 0 {
		t.Fatalf("Expected 0 port states, got %d instead\n", len(res[0].PortStates))
	}
}
