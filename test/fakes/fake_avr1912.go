package fakes

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/seanrees/denon-api/internal/util"
)

type FakeAvr1912 struct {
	ListenHost string
	ListenPort int

	// Commands to support and the response to offer. For example:
	//   "MV40": []string{"MV40", "MVMAX 800"}
	//
	// An empty value means "no response."
	CommandResponse map[string][]string

	Heartbeat string

	nl net.Listener
}

func (f *FakeAvr1912) Serve() {
	ready := make(chan bool)

	go func() {
		l := fmt.Sprintf("%s:%d", f.ListenHost, f.ListenPort)
		s, err := net.Listen("tcp", l)
		if err != nil {
			log.Fatalf("could not setup listener socket on %s: %v", l, err)
		}
		defer s.Close()

		f.nl = s
		log.Printf("listening on %v", s.Addr())

		ready <- true
		welcome := "Fake AVR1912 Telnet server"

		for {
			conn, err := s.Accept()
			if err != nil {
				log.Printf("could not accept new connection: %v (ignore this if shutting down)", err)
				break
			}

			conn.Write(util.ToBytes(welcome))

			go f.handle(conn)
		}
	}()

	<-ready
}

func (f *FakeAvr1912) Close() {
	f.nl.Close()
}

func (f *FakeAvr1912) Addr() net.Addr {
	return f.nl.Addr()
}

func (f *FakeAvr1912) handle(conn net.Conn) {
	crd := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		line, err := crd.ReadString('\r')
		if err != nil {
			if err == io.EOF {
				break
			}

			if len(f.Heartbeat) > 0 {
				conn.Write(util.ToBytes(f.Heartbeat))
			}
		}

		command := strings.TrimRight(line, "\r")

		log.Printf("received command: %s", command)

		if resp, ok := f.CommandResponse[command]; ok {
			for _, r := range resp {
				buf := util.ToBytes(r)
				out, err := conn.Write(buf)
				if err != nil || out < len(buf) {
					log.Printf("error: could not write %d bytes to %s: %v", len(buf), conn.RemoteAddr(), err)
					break
				}
			}
		}
	}
}
