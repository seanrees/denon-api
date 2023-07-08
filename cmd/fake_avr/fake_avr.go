package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/seanrees/denon-api/test/fakes"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stderr)

	f := &fakes.FakeAvr1912{
		ListenHost: "localhost",
		CommandResponse: map[string][]string{
			"MV?": {"MV40", "MVMAX 800"},
			"PW?": {"PWON"},
		},
	}
	defer f.Close()

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	f.Serve()
	<-s
}
