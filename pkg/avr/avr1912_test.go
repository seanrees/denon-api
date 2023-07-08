package avr

import "testing"

func TestParseMv(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{{
		in:   "410",
		want: "39.0",
	}, {
		in:   "39",
		want: "41.0",
	}, {
		in:   "0",
		want: "80.0",
	}, {
		in:   "80.0",
		want: "0.0",
	}, {
		in:   "95.0",
		want: "-15.0",
	}, {
		want: "-0.1",
	}}

	for _, tc := range cases {
		got := parseMv(tc.in)
		if got != tc.want {
			t.Errorf("testParseMv(%q) got %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestInvertVolume(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{{
		in:   "41.0",
		want: 390,
	}, {
		in:   "41.5",
		want: 385,
	}, {
		in:   "40",
		want: 400,
	}, {
		want: -1,
	}}

	for _, tc := range cases {
		got := invertVolume(tc.in)
		if got != tc.want {
			t.Errorf("testInvertVolume(%q) got %d want %d", tc.in, got, tc.want)
		}
	}
}
