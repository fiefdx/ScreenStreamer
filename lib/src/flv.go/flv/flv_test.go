package flv

import (
	"bytes"
	"testing"
)

func TestWriteFrame(t *testing.T) {
	got := new(bytes.Buffer)
	cFrame := CFrame{
		Stream: 55,
		Dts:    123123,
		Type:   TAG_TYPE_META,
		Body:   []byte{0x12, 0x34, 0x56, 0x78, 0x90},
	}
	err := cFrame.WriteFrame(got)
	if err != nil {
		t.Errorf("test for %s error: %s", "CFrame", err)
	} else {
		expect := []byte{
			0x12, 0x00, 0x00, 0x05, 0x01, 0xe0, 0xf3, 0x00,
			0x00, 0x00, 0x37, 0x12, 0x34, 0x56, 0x78, 0x90,
			0x00, 0x00, 0x00, 0x10,
		}
		if !bytes.Equal(expect, got.Bytes()) {
			t.Errorf("expect %x got %x", expect, got)
		}
	}
}
