package rtnetlink

import (
	"bytes"
	"net"
	"reflect"
	"testing"

	"golang.org/x/sys/unix"
)

func TestAddressMessageMarshalBinary(t *testing.T) {
	tests := []struct {
		name string
		m    Message
		b    []byte
		err  error
	}{
		{
			name: "empty",
			m:    &AddressMessage{},
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x04, 0x00, 0x01, 0x00, 0x04, 0x00, 0x02, 0x00,
				0x04, 0x00, 0x04, 0x00, 0x04, 0x00, 0x05, 0x00,
				0x04, 0x00, 0x07, 0x00, 0x08, 0x00, 0x08, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "no attributes",
			m: &AddressMessage{
				Family:       unix.AF_INET,
				PrefixLength: 8,
				Scope:        0,
				Index:        1,
				Flags:        0,
			},
			b: []byte{
				0x02, 0x08, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x04, 0x00, 0x01, 0x00, 0x04, 0x00, 0x02, 0x00,
				0x04, 0x00, 0x04, 0x00, 0x04, 0x00, 0x05, 0x00,
				0x04, 0x00, 0x07, 0x00, 0x08, 0x00, 0x08, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "attributes",
			m: &AddressMessage{
				Attributes: AddressAttributes{
					Address:   []byte{0, 0, 0, 0, 0, 0},
					Broadcast: []byte{0, 0, 0, 0, 0, 0},
					Label:     "lo",
				},
			},
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x02, 0x00,
				0x0a, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x05, 0x00,
				0x04, 0x00, 0x07, 0x00, 0x08, 0x00, 0x08, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := tt.m.MarshalBinary()

			if want, got := tt.err, err; want != got {
				t.Fatalf("unexpected error:\n- want: %v\n-  got: %v", want, got)
			}
			if err != nil {
				return
			}

			if want, got := tt.b, b; !bytes.Equal(want, got) {
				t.Fatalf("unexpected Message bytes:\n- want: [%# x]\n-  got: [%# x]", want, got)
			}
		})
	}
}

func TestAddressMessageUnmarshalBinary(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
		m    Message
		err  error
	}{
		{
			name: "empty",
			err:  errInvalidAddressMessage,
		},
		{
			name: "short",
			b:    make([]byte, 3),
			err:  errInvalidAddressMessage,
		},
		{
			name: "invalid attr",
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x04, 0x00, 0x01, 0x00, 0x04, 0x00, 0x02, 0x00,
				0x05, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			err: errInvalidAddressMessageAttr,
		},
		{
			name: "data",
			b: []byte{
				0x02, 0x08, 0xfe, 0x01, 0x01, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x01, 0x00, 0x7f, 0x00, 0x00, 0x01,
				0x08, 0x00, 0x02, 0x00, 0x7f, 0x00, 0x00, 0x01,
				0x07, 0x00, 0x03, 0x00, 0x6c, 0x6f, 0x00, 0x00,
				0x08, 0x00, 0x08, 0x00, 0x80, 0x00, 0x00, 0x00,
				0x14, 0x00, 0x06, 0x00, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0x44, 0x01, 0x00,
				0x00, 0x44, 0x01, 0x00, 0x00,
			},
			m: &AddressMessage{
				Family:       2,
				PrefixLength: 8,
				Flags:        0xfe,
				Scope:        1,
				Index:        1,
				Attributes: AddressAttributes{
					Address:   net.IP{0x7f, 0x0, 0x0, 0x1},
					Local:     net.IP{0x7f, 0x0, 0x0, 0x1},
					Label:     "lo",
					Broadcast: net.IP(nil),
					Anycast:   net.IP(nil),
					CacheInfo: CacheInfo{
						Prefered: 0xffffffff,
						Valid:    0xffffffff,
						Created:  0x144,
						Updated:  0x144,
					},
					Multicast: net.IP(nil),
					Flags:     0x80,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &AddressMessage{}
			err := (m).UnmarshalBinary(tt.b)

			if want, got := tt.err, err; want != got {
				t.Fatalf("unexpected error:\n- want: %v\n-  got: %v", want, got)
			}
			if err != nil {
				return
			}

			if want, got := tt.m, m; !reflect.DeepEqual(want, got) {
				t.Fatalf("unexpected Message:\n- want: %#v\n-  got: %#v", want, got)
			}
		})
	}
}
