package echo

import (
	"context"
	"net"
	"os"
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

func datagramEchoServer(ctx context.Context, network, addr string) (net.Addr, error) {
	socket, err := net.ListenPacket(network, addr)
	if err != nil {
		return nil, err
	}

	go func() {
		go func() {
			<-ctx.Done()
			socket.Close()
			if network == "unixgram" {
				os.Remove(addr)
			}
		}()

		buf := make([]byte, 1024)
		for {
			n, clientAddr, err := socket.ReadFrom(buf)
			if err != nil {
				return
			}

			if _, err := socket.WriteTo(buf[:n], clientAddr); err != nil {
				return
			}
		}
	}()

	return socket.LocalAddr(), nil
}
