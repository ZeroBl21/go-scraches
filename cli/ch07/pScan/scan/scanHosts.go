// Package scan provides types and functions to perform TCP port
// scans on a list of hosts
package scan

import (
	"fmt"
	"net"
	"time"
)

const (
	PORT_OPEN   = "open"
	PORT_CLOSED = "closed"
)

// PortState represends the state of a single TCP port
type PortState struct {
	Port int
	Open state
}

type state bool

func (s state) String() string {
	if s {
		return PORT_OPEN
	}

	return PORT_CLOSED
}

// scanPort performs a port scan on  a single TCP port
func scanPort(host string, port int, timeout time.Duration) PortState {
	p := PortState{
		Port: port,
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	scanConn, err := net.DialTimeout("tcp", addr, timeout*time.Millisecond)
	if err != nil {
		return p
	}
	scanConn.Close()
	p.Open = true

	return p
}

// Result represends the scan results for a single host
type Results struct {
	Host       string
	NotFound   bool
	PortStates []PortState
}

// Run performs a port scan on the hosts list
func Run(hl *HostsList, ports []int, timeout time.Duration) []Results {
	res := make([]Results, 0, len(hl.Hosts))
	for _, h := range hl.Hosts {
		r := Results{
			Host: h,
		}

		if _, err := net.LookupHost(h); err != nil {
			r.NotFound = true
			res = append(res, r)
			continue
		}

		for _, p := range ports {
			r.PortStates = append(r.PortStates, scanPort(h, p, timeout))
		}

		res = append(res, r)
	}

	return res
}
