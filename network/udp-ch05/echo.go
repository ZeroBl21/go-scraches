package echo

import (
	"context"
	"fmt"
	"net"
)

func echoServerUDP(ctx context.Context, addr string) (net.Addr, error) {
	socket, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("binding to udp %s: %w", addr, err)
	}

	go func() {
		go func() {
			<-ctx.Done()
			socket.Close()
		}()

		buf := make([]byte, 1024)
		for {
			n, clientAddr, err := socket.ReadFrom(buf)
			if err != nil {
				return
			}

			_, err = socket.WriteTo(buf[:n], clientAddr)
			if err != nil {
				return
			}
		}
	}()

	return socket.LocalAddr(), nil
}
