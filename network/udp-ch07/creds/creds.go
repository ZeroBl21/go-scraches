package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"

	"github.com/ZeroBl21/go-network/ch07/creds/auth"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage:\n\t%s <group names>\n",
			filepath.Base(os.Args[0]))

		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	groups := parseGroupNames(flag.Args())
	socket := filepath.Join(os.TempDir(), "creds.sock")
	addr, err := net.ResolveUnixAddr("unix", socket)
	if err != nil {
		log.Fatal(err)
	}

	ss, err := net.ListenUnix("unix", addr)
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		ss.Close()
	}()

	fmt.Printf("Listening on %s ...\n", socket)

	for {
		conn, err := ss.AcceptUnix()
		if err != nil {
			break
		}

		if !auth.Allowed(conn, groups) {
			if _, err = conn.Write([]byte("Access denied\n")); err != nil {
				log.Println(err)
				conn.Close()
				continue
			}
		}

		if _, err = conn.Write([]byte("Welcome\n")); err != nil {
			log.Println(err)
			conn.Close()
		}
	}
}

func parseGroupNames(args []string) map[string]struct{} {
	groups := make(map[string]struct{})

	for _, arg := range args {
		group, err := user.LookupGroup(arg)
		if err != nil {
			log.Println(err)
			continue
		}

		groups[group.Gid] = struct{}{}
	}

	return groups
}
