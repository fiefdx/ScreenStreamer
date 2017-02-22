package flv


import (
    "bytes"
    "fmt"
    "encoding/binary"
)

var (
    be = binary.BigEndian
)

type BitReader struct {
    reader      *bytes.Reader
    bitBuffer   uint32
    bitsInBuf   uint32
}

func NewBitReader(buffer []byte) (r *BitReader) {
    reader := bytes.NewReader(buffer)
    r = &BitReader{reader: reader}
    return
}

func (r *BitReader) readByte() (result uint8) {
    result, err := r.reader.ReadByte()
    if err != nil {
        panic(fmt.Errorf("BitReader.readByte: %s", err))
    }
    return result
}

func (r *BitReader) Seek(offset int64, whence int) {
    _, err := r.reader.Seek(offset, whence)
    if err != nil {
        panic(fmt.Errorf("BitReader.Seek: %s", err))
    }
    r.bitBuffer = 0
    r.bitsInBuf = 0
}

func (r *BitReader) Read(b []byte) (n int) {
    n, err := r.reader.Read(b)
    if err != nil {
        panic(fmt.Errorf("BitReader.Read: %s", err))
    }
    return
}

func (r *BitReader) readBits(count uint32) (result uint32) {
    if count > 32 {
        panic(fmt.Errorf("BitReader.readBits: count = %d but should be less than 32", count))
    }
    for count > r.bitsInBuf {
        r.bitBuffer <<= 8
        r.bitBuffer |= uint32(r.readByte())
        r.bitsInBuf += 8

        if r.bitsInBuf > 24 {
            if count <= r.bitsInBuf {
                break
            }
            if count <= 32 {
                return r.readBits(16) << 16 | r.readBits(count-16)
            }
        }
    }
    r.bitsInBuf -= count
    return (r.bitBuffer >> r.bitsInBuf) & ((uint32(1) << count)-1)
}

func (r *BitReader) U(count uint32) (uint32) {
    return r.readBits(count)
}

func (r *BitReader) U8() (byte) {
    return byte(r.U(8))
}

func (r *BitReader) Ue() (result uint32) {
    leadingZeroes := uint32(0)
    for r.readBits(1) == 0 {
        leadingZeroes += 1
    }
    if leadingZeroes == 0 {
        return 0
    }
    if leadingZeroes >= 32 {
        panic(fmt.Errorf("BitReader.Ue overflow, leadingZeroes = %d", leadingZeroes))
    }
    remaining := r.readBits(leadingZeroes)
    return (1<<leadingZeroes-1+remaining)
}

func (r *BitReader) Se() (int32) {
    t := r.Ue() + 1
    if t > 0xFFFFFFFE {
        panic(fmt.Errorf("BitReader.Se overflow"))
    }
    return int32((1 - 2 * (t & 1)) * (t / 2));
}