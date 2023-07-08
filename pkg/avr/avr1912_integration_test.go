package avr

import (
	"regexp"
	"testing"
	"time"

	"github.com/seanrees/denon-api/test/fakes"
)

func TestUnrequestedEvents(t *testing.T) {
	f := fakes.FakeAvr1912{
		CommandResponse: map[string][]string{
			"MS?": {"MSFOO"},
		},
		Heartbeat: "MSBAZ",
	}
	f.Serve()
	t.Cleanup(f.Close)

	d := NewAvr1912(f.Addr().String())
	defer d.Close()

	time.Sleep(150 * time.Millisecond)
	ch := make(chan []string)
	r := request{command: "MS?", expect: regexp.MustCompile("MS.*"), response: ch}
	defer close(r.response)

	d.conn.commands <- r
	resp := <-ch

	if got := len(resp); got != 1 {
		t.Errorf("len(resp) got %d, want 1", got)
	}
	want := "MSFOO"
	if got := resp[0]; got != want {
		t.Errorf("resp: got %s want %s", got, want)
	}
}

func TestTimeout(t *testing.T) {
	old := Avr1912CommandTimeout
	t.Cleanup(func() {
		Avr1912CommandTimeout = old
	})
	Avr1912CommandTimeout = 1 * time.Millisecond

	f := fakes.FakeAvr1912{
		CommandResponse: map[string][]string{
			"PW?": {},
		},
	}
	f.Serve()
	t.Cleanup(f.Close)

	d := NewAvr1912(f.Addr().String())
	defer d.Close()

	want := "FAILED"
	if got := d.Power(); got != want {
		t.Errorf("Power(), got %s want %s", got, want)
	}
}

func TestReconnect(t *testing.T) {
	old := Avr1912ReconnectTimeout
	t.Cleanup(func() {
		Avr1912ReconnectTimeout = old
	})
	Avr1912ReconnectTimeout = 1 * time.Millisecond

	f1 := fakes.FakeAvr1912{
		CommandResponse: map[string][]string{
			"SI?": {"SISHOULD NOT SEE"},
		},
	}
	f1.Serve()

	d := NewAvr1912(f1.Addr().String())
	defer d.Close()
	f1.Close()

	f2 := fakes.FakeAvr1912{
		CommandResponse: map[string][]string{
			"SI?": {"SISHOULD SEE"},
		},
	}
	f2.Serve()
	defer f2.Close()

	// Hack
	d.conn.addr = f2.Addr().String()

	want := "SHOULD SEE"
	if got := d.InputSource(); got != want {
		t.Errorf("InputSource(), got %s want %s", got, want)
	}
}

func TestInputSource(t *testing.T) {
	f := fakes.FakeAvr1912{
		CommandResponse: map[string][]string{
			"SI?":   {"SIFOO"},
			"SIBAR": {"SIBAR"},
		},
	}
	f.Serve()
	t.Cleanup(f.Close)

	d := NewAvr1912(f.Addr().String())
	defer d.Close()

	want := "FOO"
	if got := d.InputSource(); got != want {
		t.Errorf("InputSource(), got %s want %s", got, want)
	}

	want = "BAR"
	if got := d.SetInputSource(want); got != want {
		t.Errorf("SetInputSource(), got %s want %s", got, want)
	}
}

func TestPower(t *testing.T) {
	f := fakes.FakeAvr1912{
		CommandResponse: map[string][]string{
			"PW?":       {"PWON"},
			"PWSTANDBY": {"PWSTANDBY", "IGNORE"},
			"PWON":      {"PWON"},
		},
	}
	f.Serve()
	t.Cleanup(f.Close)

	d := NewAvr1912(f.Addr().String())
	defer d.Close()

	want := "ON"
	if got := d.Power(); got != want {
		t.Errorf("Power(), got %s want %s", got, want)
	}

	want = "STANDBY"
	if got := d.Standby(); got != want {
		t.Errorf("Standby(), got %s want %s", got, want)
	}

	want = "ON"
	if got := d.PowerOn(); got != want {
		t.Errorf("PowerOn(), got %s want %s", got, want)
	}
}

func TestSurroundMode(t *testing.T) {
	f := fakes.FakeAvr1912{
		CommandResponse: map[string][]string{
			"MS?":   {"MSFOO"},
			"MSBAR": {"MSBAR"},
		},
	}
	f.Serve()
	t.Cleanup(f.Close)

	d := NewAvr1912(f.Addr().String())
	defer d.Close()

	want := "FOO"
	if got := d.SurroundMode(); got != want {
		t.Errorf("SurroundMode(), got %s want %s", got, want)
	}

	want = "BAR"
	if got := d.SetSurroundMode(want); got != want {
		t.Errorf("SetSurroundMode(), got %s want %s", got, want)
	}
}

func TestVolume(t *testing.T) {
	f := fakes.FakeAvr1912{
		CommandResponse: map[string][]string{
			"MV?":    {"MV40", "MVMAX 800"},
			"MVUP":   {"MV395"},
			"MVDOWN": {"MV400"},
			"MV505":  {"MV505", "MVMAX 800"},
		},
	}
	f.Serve()
	t.Cleanup(f.Close)

	d := NewAvr1912(f.Addr().String())
	defer d.Close()

	want := "40.0"
	if got := d.Volume(); got != want {
		t.Errorf("Volume(), got %s want %s", got, want)
	}

	want = "40.5"
	if got := d.VolumeUp(); got != want {
		t.Errorf("VolumeUp(), got %s want %s", got, want)
	}

	want = "40.0"
	if got := d.VolumeDown(); got != want {
		t.Errorf("VolumeUp(), got %s want %s", got, want)
	}

	want = "29.5"
	if got := d.SetVolume(want); got != want {
		t.Errorf("SetVolume(%q), got %s want %s", want, got, want)
	}
}
