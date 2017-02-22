package amf0

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
)

type Encoder struct {
	w       io.Writer
	bw      *bufio.Writer
	refObjs []interface{}
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w, bw: bufio.NewWriter(w)}
}

func (enc *Encoder) Encode(v interface{}) error {
	err := enc.encodeValue(v)
	if err != nil {
		return err
	}
	err = enc.bw.Flush()
	return err
}

func (enc *Encoder) encodeValue(v interface{}) error {
	u32 := make([]byte, 4)
	u64 := make([]byte, 8)
	if value, ok := v.(NumberType); ok {
		err := enc.bw.WriteByte(NumberMarker)
		if err != nil {
			return err
		}
		number := math.Float64bits(float64(value))
		binary.BigEndian.PutUint64(u64, number)
		_, err = enc.bw.Write(u64)
		if err != nil {
			return err
		}
	} else if value, ok := v.(BooleanType); ok {
		err := enc.bw.WriteByte(BooleanMarker)
		if err != nil {
			return err
		}
		if value {
			err = enc.bw.WriteByte(1)
		} else {
			err = enc.bw.WriteByte(0)
		}
		if err != nil {
			return err
		}
	} else if value, ok := v.(StringType); ok {
		err := enc.bw.WriteByte(StringMarker)
		if err != nil {
			return err
		}
		err = writeUTF8(enc.bw, value)
		if err != nil {
			return err
		}
	} else if value, ok := v.(*ObjectType); ok {
		ok, err := enc.writeRef(value)
		if err != nil {
			return err
		}
		if ok {
			return nil
		} else {
			enc.refObjs = append(enc.refObjs, value)
			err := enc.bw.WriteByte(ObjectMarker)
			if err != nil {
				return err
			}
			err = enc.writeObject(_Object(*value))
			if err != nil {
				return err
			}
		}
	} else if _, ok := v.(NullType); ok {
		err := enc.bw.WriteByte(NullMarker)
		if err != nil {
			return err
		}
	} else if _, ok := v.(UndefinedType); ok {
		err := enc.bw.WriteByte(UndefinedMarker)
		if err != nil {
			return err
		}
	} else if value, ok := v.(*EcmaArrayType); ok {
		ok, err := enc.writeRef(value)
		if err != nil {
			return err
		}
		if ok {
			return nil
		} else {
			enc.refObjs = append(enc.refObjs, value)
			err := enc.bw.WriteByte(EcmaArrayMarker)
			if err != nil {
				return err
			}
			associativeCount := len(*value)
			binary.BigEndian.PutUint32(u32, uint32(associativeCount))
			_, err = enc.bw.Write(u32)
			if err != nil {
				return err
			}
			err = enc.writeObject(_Object(*value))
			if err != nil {
				return err
			}
		}
	} else if value, ok := v.(*StrictArrayType); ok {
		ok, err := enc.writeRef(value)
		if err != nil {
			return err
		}
		if ok {
			return nil
		} else {
			enc.refObjs = append(enc.refObjs, value)
			err := enc.bw.WriteByte(StrictArrayMarker)
			if err != nil {
				return err
			}
			arrayCount := len(*value)
			binary.BigEndian.PutUint32(u32, uint32(arrayCount))
			_, err = enc.bw.Write(u32)
			if err != nil {
				return err
			}
			for i := 0; i < arrayCount; i++ {
				err := enc.encodeValue((*value)[i])
				if err != nil {
					return err
				}
			}
		}
	} else if value, ok := v.(DateType); ok {
		err := enc.bw.WriteByte(DateMarker)
		if err != nil {
			return err
		}
		date := math.Float64bits(value.Date)
		binary.BigEndian.PutUint64(u64, date)
		_, err = enc.bw.Write(u64)
		if err != nil {
			return err
		}
		_, err = enc.bw.Write([]byte{0x00, 0x00})
		if err != nil {
			return err
		}
	} else if _, ok := v.(LongStringType); ok {
		err := enc.bw.WriteByte(LongStringMarker)
		if err != nil {
			return err
		}
		err = writeUTF8Long(enc.bw, v.(LongStringType))
		if err != nil {
			return err
		}
	} else if _, ok := v.(UnsupportedType); ok {
		err := enc.bw.WriteByte(UnsupportedMarker)
		if err != nil {
			return err
		}
	} else if value, ok := v.(XmlDocumentType); ok {
		err := enc.bw.WriteByte(XmlDocumentMarker)
		if err != nil {
			return err
		}
		err = writeUTF8Long(enc.bw, LongStringType(value))
		if err != nil {
			return err
		}
	} else if value, ok := v.(*TypedObjectType); ok {
		ok, err := enc.writeRef(value)
		if err != nil {
			return err
		}
		if !ok {
			enc.refObjs = append(enc.refObjs, value)
			err := enc.bw.WriteByte(TypedObjectMarker)
			if err != nil {
				return err
			}
			err = writeUTF8(enc.bw, value.ClassName)
			if err != nil {
				return err
			}
			err = enc.writeObject(value.Object)
			if err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("amf0 encoder unsupported type: %v", reflect.TypeOf(v))
	}
	return nil
}

func (enc *Encoder) writeRef(v interface{}) (bool, error) {
	u16 := make([]byte, 2)
	for i, obj := range enc.refObjs {
		if v == obj {
			err := enc.bw.WriteByte(ReferenceMarker)
			if err != nil {
				return false, err
			}
			binary.BigEndian.PutUint16(u16, uint16(i))
			_, err = enc.bw.Write(u16)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func (enc *Encoder) writeObject(obj _Object) error {
	for k, v := range obj {
		err := writeUTF8(enc.bw, k)
		if err != nil {
			return err
		}
		err = enc.encodeValue(v)
		if err != nil {
			return err
		}
	}
	_, err := enc.bw.Write([]byte{0x00, 0x00, ObjectEndMarker})
	if err != nil {
		return err
	}
	return nil
}

func writeUTF8(w io.Writer, s StringType) error {
	u16 := make([]byte, 2)
	length := len(s)
	if length > 0xFFFF {
		return errors.New("string too long")
	}
	binary.BigEndian.PutUint16(u16, uint16(length))
	_, err := w.Write(u16)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	if err != nil {
		return err
	}
	return nil
}

func writeUTF8Long(w io.Writer, s LongStringType) error {
	u32 := make([]byte, 4)
	length := len(s)
	binary.BigEndian.PutUint32(u32, uint32(length))
	_, err := w.Write(u32)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	if err != nil {
		return err
	}
	return nil
}
