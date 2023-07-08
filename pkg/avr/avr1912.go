package avr

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// How long to wait for a command to return, before returning an empty
	// string. Experimentally, some commands like switching modes, can take
	// 1 to 2 seconds.
	Avr1912CommandTimeout = 2 * time.Second

	// Time to wait before reconnecting, after disconnect.
	Avr1912ReconnectTimeout = 10 * time.Second
)

type DenonAvr1912 struct {
	conn *avr1912Connection
}

type avr1912Connection struct {
	InputSource  string
	MasterVolume string
	Power        string
	SurroundMode string

	// Address (host:port) to connect to a Denon AVR1912.
	addr string

	// Command queue (in Denon-ese) for the AVR.
	commands chan request

	// Send true to shut this connection down cleanly.
	cancel chan bool
}

type request struct {
	command  string
	expect   *regexp.Regexp
	response chan<- []string
}

func NewAvr1912(addr string) *DenonAvr1912 {
	r := DenonAvr1912{
		conn: &avr1912Connection{
			addr:     addr,
			commands: make(chan request),
			cancel:   make(chan bool),
		},
	}

	go r.conn.connect()
	return &r
}

func (avr *avr1912Connection) connect() {
	for {
		clean := avr.innerConnect()
		if clean {
			break
		}

		log.Printf("received disconnect, waiting %v seconds to reconnect", Avr1912ReconnectTimeout)
		time.Sleep(Avr1912ReconnectTimeout)
	}
}

func (avr *avr1912Connection) innerConnect() bool {
	conn, err := net.DialTimeout("tcp", avr.addr, 2*time.Second)
	if err != nil {
		log.Printf("error: could not dial %s: %v", avr.addr, err)
		return false
	}

	log.Printf("connected to %s", avr.addr)

	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Printf("warn: could not close connection to %s: %v", avr.addr, err)
		}
	}()

	crd := bufio.NewReader(conn)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case c := <-avr.commands:
			log.Printf("writing %s", c.command)
			buf := make([]byte, len(c.command)+1)
			copy(buf, c.command)
			buf[len(c.command)] = '\r'

			out, err := conn.Write(buf)
			if err != nil || out < len(buf) {
				log.Printf("error: could not write %d bytes to %s: %v", len(buf), avr.addr, err)
				return false
			}

			deadline := time.Now().Add(Avr1912CommandTimeout)
			responded := false
			for time.Now().Before(deadline) {
				conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
				resp, err := read(crd, c.expect)
				if err != nil {
					return false
				}
				if len(resp) > 0 {
					log.Printf("responding with %v", resp)
					c.response <- resp
					responded = true
					break
				}
			}
			if !responded {
				c.response <- []string{}
			}
		case <-ticker.C:
			conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
			resp, err := read(crd, regexp.MustCompile(".*"))

			for _, e := range resp {
				log.Printf("unexpected event: %s (ignoring)", e)
			}

			if err != nil {
				return false
			}
		case <-avr.cancel:
			log.Printf("received cancelation signal, shutting down cleanly")
			return true
		}
	}
}

func read(rd *bufio.Reader, expect *regexp.Regexp) ([]string, error) {
	// Read out the responses. These might be ~immediate (10's of ms) or ~slow
	// (1-2 secs).
	resp := []string{}

	for {
		line, err := rd.ReadString('\r')
		if err != nil {
			if err == io.EOF {
				return resp, err
			}

			// Otherwise it's just an i/o timeout.
			break
		}

		event := strings.TrimRight(line, "\r")
		if expect.MatchString(event) {
			resp = append(resp, event)
		} else {
			log.Printf("unrequested response: %s (ignoring)", event)
		}
	}

	return resp, nil
}

func (avr *avr1912Connection) close() {
	avr.cancel <- true
	close(avr.cancel)
	close(avr.commands)
}

func (avr *avr1912Connection) sendAndWait(c string, e *regexp.Regexp) string {
	ch := make(chan []string)

	r := request{command: c, expect: e, response: ch}
	defer close(r.response)

	avr.commands <- r
	resp := <-ch

	log.Printf("response %v", resp)

	if len(resp) == 0 {
		return "XXFAILED"
	} else {
		return resp[len(resp)-1]
	}
}

func (avr *avr1912Connection) ms(sub string) string {
	resp := avr.sendAndWait("MS"+sub, regexp.MustCompile("MS.*"))
	return resp[2:]
}

func (avr *avr1912Connection) mv(sub string) string {
	resp := avr.sendAndWait("MV"+sub, regexp.MustCompile("MV[0-9]+"))
	return resp[2:]
}

func (avr *avr1912Connection) pw(sub string) string {
	resp := avr.sendAndWait("PW"+sub, regexp.MustCompile("PW.*"))
	return resp[2:]
}

func (avr *avr1912Connection) si(sub string) string {
	resp := avr.sendAndWait("SI"+sub, regexp.MustCompile("SI.*"))
	return resp[2:]
}

func (api *DenonAvr1912) InputSource() string {
	return api.conn.si("?")
}

func (api *DenonAvr1912) SetInputSource(input string) string {
	return api.conn.si(input)
}

func (api *DenonAvr1912) Power() string {
	return api.conn.pw("?")
}

func (api *DenonAvr1912) PowerOn() string {
	return api.conn.pw("ON")
}

func (api *DenonAvr1912) Standby() string {
	return api.conn.pw("STANDBY")
}

func (api *DenonAvr1912) SurroundMode() string {
	return api.conn.ms("?")
}

func (api *DenonAvr1912) SetSurroundMode(mode string) string {
	return api.conn.ms(mode)
}

func (api *DenonAvr1912) Volume() string {
	return parseMv(api.conn.mv("?"))
}

func (api *DenonAvr1912) VolumeUp() string {
	return parseMv(api.conn.mv("UP"))
}
func (api *DenonAvr1912) VolumeDown() string {
	return parseMv(api.conn.mv("DOWN"))
}

func (api *DenonAvr1912) SetVolume(level string) string {
	return parseMv(api.conn.mv(strconv.Itoa(invertVolume(level))))
}

func (api *DenonAvr1912) Close() {
	api.conn.close()
	api.conn = nil
}

// Takes a decimal dB level (e.g; 39.5) and converts it to the format to/from Denon.
func invertVolume(level string) int {
	f, err := strconv.ParseFloat(level, 32)
	if err != nil {
		log.Printf("could not parse %s: %v", level, err)
		return -1
	}

	// Denon uses an inverse scale in the API from what displays on the unit.
	// The range is [0, 80]. The unit displays 0+delta, while the API uses
	// 80-delta. Further, it only operates in 0.5 level increments.
	//
	// We try to expose the same interface in our API as on the unit itavr,
	// so humans can reason about it.
	i := int(800 - f*10)
	i = i - (i % 5)

	return i
}

// Handles Denon-ese (e.g; 39, 395, 410, etc.)
func parseMv(mv string) string {
	if len(mv) == 3 {
		// E.g: 415 -> 41.5
		mv = fmt.Sprintf("%s.%s", mv[0:2], mv[2:])
	}

	// Convert back to decimal for human use
	vol := float32(invertVolume(mv)) / 10
	return fmt.Sprintf("%.1f", vol)
}
