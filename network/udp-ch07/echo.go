package echo

import (
	"context"
	"net"
)

func streamingEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
	socket, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}

	go func() {
		go func() {
			<-ctx.Done()
			socket.Close()
		}()

		for {
			conn, err := socket.Accept()
			if err != nil {
				return
			}

			go func() {
				defer conn.Close()

				for {
					buf := make([]byte, 1024)
					n, err := conn.Read(buf)
					if err != nil {
						return
					}

					_, err = conn.Write(buf[:n])
					if err != nil {
						return
					}
				}
			}()
		}
	}()

	return socket.Addr(), nil
}
