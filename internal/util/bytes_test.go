package util

import (
	"reflect"
	"testing"
)

func TestToBytes(t *testing.T) {
	cases := []struct {
		in   string
		want []byte
	}{{
		want: []byte{'\r'},
	}, {
		in:   "foo",
		want: []byte{'f', 'o', 'o', '\r'},
	}, {
		in:   "test\r",
		want: []byte{'t', 'e', 's', 't', '\r', '\r'},
	}, {
		in:   "sp a ce",
		want: []byte{'s', 'p', ' ', 'a', ' ', 'c', 'e', '\r'},
	}}

	for _, tc := range cases {
		got := ToBytes(tc.in)

		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("ToBytes(%q) got %q want %q", tc.in, got, tc.want)
		}
	}
}
