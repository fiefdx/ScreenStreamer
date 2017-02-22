package flv;

import (
    "testing"
    // "bytes"
)

func TestBitReader(t *testing.T) {
    data1 := []byte {0xF, 0xFF, (1<<7)|(011), (1<<7)|(011)};
    r := NewBitReader(data1)

    p1 := r.U(4)
    if p1 != 0 {
        t.Errorf("U: 0 != %d", p1)
    }

    p2 := r.U(12)
    if p2 != 0xFFF {
        t.Errorf("U: 0xFFF != %d", p1)
    }

    ue1 := r.Ue()
    if ue1 != 0 {
        t.Errorf("Ue: 0 != %d", ue1)
    }

    ue2 := r.Ue()
    if ue2 != 8 {
        t.Errorf("Ue: 8 != %d", ue2)
    }

    se1 := r.Se()
    if se1 != 0 {
        t.Errorf("Se: 0 != %d", se1)
    }

    se2 := r.Se()
    if se2 != -4 {
        t.Errorf("Se: -4 != %d", se2)
    }
}