package amf3

import (
	"errors"
	"io"
)

func encodeUInt29(n uint32) ([]byte, error) {
	var b []byte
	if n <= 0x0000007F {
		b = make([]byte, 1)
		b[0] = byte(n)
	} else if n <= 0x00003FFF {
		b = make([]byte, 2)
		b[0] = byte(n>>7 | 0x80)
		b[1] = byte(n & 0x7F)
	} else if n <= 0x001FFFFF {
		b = make([]byte, 3)
		b[0] = byte(n>>14 | 0x80)
		b[1] = byte(n>>7&0x7F | 0x80)
		b[2] = byte(n & 0x7F)
	} else if n <= 0x1FFFFFFF {
		b = make([]byte, 4)
		b[0] = byte(n>>22 | 0x80)
		b[1] = byte(n>>15&0x7F | 0x80)
		b[2] = byte(n>>8&0x7F | 0x80)
		b[3] = byte(n)
	} else {
		return nil, errors.New("out of range")
	}
	return b, nil
}

func EncodeUInt29(w io.Writer, n uint32) error {
	b, err := encodeUInt29(n)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func EncodeInt29(w io.Writer, n int32) error {
	un, err := S2UInt29(n)
	if err != nil {
		return err
	}
	un = un&0xFFFFFFF | (un & 0x80000000 >> 3)
	return EncodeUInt29(w, un)
}

func DecodeUInt29(r io.Reader) (uint32, error) {
	var n uint32 = 0
	i := 0
	b := make([]byte, 1)
	for {
		_, err := r.Read(b)
		if err != nil {
			return 0, err
		}
		if i != 3 {
			n |= uint32(b[0] & 0x7F)
			if b[0]&0x80 != 0 {
				if i != 2 {
					n <<= 7
				} else {
					n <<= 8
				}
			} else {
				break
			}
		} else {
			n |= uint32(b[0])
			break
		}
		i++
	}
	return n, nil
}

func DecodeInt29(r io.Reader) (int32, error) {
	un, err := DecodeUInt29(r)
	if err != nil {
		return 0, err
	}
	sn, err := U2SInt29(un)
	if err != nil {
		return 0, err
	}
	return sn, nil
}

// signed int -> uint29
func S2UInt29(i int32) (uint32, error) {
	if i > 0xFFFFFFF || i < -0x10000000 {
		return 0, errors.New("out of range")
	}
	ui := uint32(i)
	ui = ui&0xFFFFFFF | (ui & 0x80000000 >> 3)
	return ui, nil
}

// uint29 -> signed int
func U2SInt29(i uint32) (int32, error) {
	if i > 0x1FFFFFFF {
		return 0, errors.New("out of range")
	}
	if i&0x10000000 != 0 {
		return int32(i | 0xFF000000), nil
	}
	return int32(i), nil
}
