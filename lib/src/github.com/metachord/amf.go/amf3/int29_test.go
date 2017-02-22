package amf3

import (
	"bytes"
	"testing"
)

type uint32BytesPair struct {
	i uint32
	b []byte
}

var testUint32BytesPair = []uint32BytesPair{
	{0x00000012, []byte{0x12}},
	{0x00001234, []byte{0xa4, 0x34}},
	{0x00123456, []byte{0xc8, 0xe8, 0x56}},
	{0x12345678, []byte{0xc8, 0xe8, 0xd6, 0x78}},
}

func TestEncodeUInt29(t *testing.T) {
	for _, pair := range testUint32BytesPair {
		b, err := encodeUInt29(pair.i)
		if err != nil {
			t.Errorf("test for %x: %s", pair.i, err)
		} else {
			if !bytes.Equal(b, pair.b) {
				t.Errorf("test for %x: expect %x got %x", pair.i, pair.b, b)
			}
		}
	}
	_, err := encodeUInt29(0xffffffff)
	if err == nil {
		t.Errorf("test for 0xffffffff: should report out of range")
	}
}

func TestDecodeUint29(t *testing.T) {
	for _, pair := range testUint32BytesPair {
		r := bytes.NewReader(pair.b)
		u29, err := DecodeUInt29(r)
		if err != nil {
			t.Errorf("test for %x: err", pair.b)
		} else {
			if u29 != pair.i {
				t.Errorf("test for %x: expect %x got %x", pair.b, pair.i, u29)
			}
		}
	}
}

type intPair struct {
	u uint32
	s int32
}

var testIntPair = []intPair{
	{0x01111111, 0x01111111},
	{0x0FFFFFFF, 0x0FFFFFFF},
	{0x1FFFFFFF, -0x00000001},
}

func TestS2UInt29(t *testing.T) {
	for _, pair := range testIntPair {
		u, err := S2UInt29(pair.s)
		if err != nil {
			t.Errorf("test for %v error", pair.s)
		} else {
			if u != pair.u {
				t.Errorf("test for %v: expect %v got %v", pair.s, pair.u, u)
			}
		}
	}
	_, err := S2UInt29(0x10000000)
	if err == nil {
		t.Errorf("test for 0x10000000: should report out of range")
	}
}

func TestU2SInt29(t *testing.T) {
	for _, pair := range testIntPair {
		s, err := U2SInt29(pair.u)
		if err != nil {
			t.Errorf("test for %v error", pair.u)
		} else {
			if s != pair.s {
				t.Errorf("test for %v: expect %v got %v", pair.u, pair.s, s)
			}
		}
	}
	_, err := U2SInt29(0xFFFFFFFF)
	if err == nil {
		t.Errorf("test for 0xFFFFFFFF: should report out of range")
	}
}
