package discovery_test

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	. "github.com/ZeroBl21/dsg/ch01/proglog/internal/discovery"
	"github.com/hashicorp/serf/serf"
	"github.com/stretchr/testify/require"
	"github.com/travisjeffery/go-dynaport"
)

func TestMembership(t *testing.T) {
	m, handler := setupMember(t, nil)
	m, _ = setupMember(t, m)
	m, _ = setupMember(t, m)

	require.Eventually(t, func() bool {
		return 2 == len(handler.joins) &&
			3 == len(m[0].Members()) &&
			0 == len(handler.leaves)
	}, 3*time.Second, 250*time.Millisecond)

	require.NoError(t, m[2].Leave())

	require.Eventually(t, func() bool {
		return 2 == len(handler.joins) &&
			3 == len(m[0].Members()) &&
			serf.StatusLeft == m[0].Members()[2].Status &&
			1 == len(handler.leaves)
	}, 3*time.Second, 250*time.Millisecond)
}

func setupMember(t *testing.T, members []*Membership) ([]*Membership, *handler) {
	id := len(members)
	ports := dynaport.Get(1)
	port := strconv.Itoa(ports[0])

	addr := net.JoinHostPort("127.0.0.1", port)
	tags := map[string]string{
		"rpc_addr": addr,
	}

	c := Config{
		NodeName: fmt.Sprintf("%d", id),
		BindAddr: addr,
		Tags:     tags,
	}

	h := &handler{}
	if len(members) == 0 {
		h.joins = make(chan map[string]string, 3)
		h.leaves = make(chan string, 3)
	} else {
		c.StartJoinAddrs = []string{
			members[0].BindAddr,
		}
	}

	m, err := New(h, c)
	require.NoError(t, err)

	members = append(members, m)

	return members, h
}

type handler struct {
	joins  chan map[string]string
	leaves chan string
}

func (h *handler) Join(id, addr string) error {
	if h.joins != nil {
		h.joins <- map[string]string{
			"id":   id,
			"addr": addr,
		}
	}

	return nil
}

func (h *handler) Leave(id string) error {
	if h.leaves != nil {
		h.leaves <- id
	}

	return nil
}
