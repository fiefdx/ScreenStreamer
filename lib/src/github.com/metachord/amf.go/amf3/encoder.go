package amf3

import (
	"bufio"
	"encoding/binary"
	"io"
	"math"
)

type Encoder struct {
	w          io.Writer
	bw         *bufio.Writer
	refStrings []StringType  // Strings
	refObjects []interface{} // Object, Array, XML, XMLDocument, ByteArray, Date and instances of user defined Classes
	refTraits  []*Trait      // Objects and instances of user defined Classes have trait information
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
	u64 := make([]byte, 8)
	if _, ok := v.(UndefinedType); ok {
		_, err := enc.bw.Write([]byte{UndefinedMarker})
		if err != nil {
			return err
		}
	} else if _, ok := v.(NullType); ok {
		_, err := enc.bw.Write([]byte{NullMarker})
		if err != nil {
			return err
		}
	} else if _, ok := v.(FalseType); ok {
		_, err := enc.bw.Write([]byte{FalseMarker})
		if err != nil {
			return err
		}
	} else if _, ok := v.(TrueType); ok {
		_, err := enc.bw.Write([]byte{TrueMarker})
		if err != nil {
			return err
		}
	} else if value, ok := v.(IntegerType); ok {
		_, err := enc.bw.Write([]byte{IntegerMarker})
		if err != nil {
			return err
		}
		err = EncodeUInt29(enc.bw, uint32(value))
		if err != nil {
			return err
		}
	} else if value, ok := v.(DoubleType); ok {
		_, err := enc.bw.Write([]byte{DoubleMarker})
		if err != nil {
			return err
		}
		number := math.Float64bits(float64(value))
		binary.BigEndian.PutUint64(u64, number)
		_, err = enc.bw.Write(u64)
		if err != nil {
			return err
		}
	} else if value, ok := v.(StringType); ok {
		_, err := enc.bw.Write([]byte{StringMarker})
		if err != nil {
			return err
		}
		err = enc.writeString(value)
		if err != nil {
			return err
		}
	} else if value, ok := v.(*XMLDocumentType); ok {
		_, err := enc.bw.Write([]byte{XmlDocMarker})
		if err != nil {
			return err
		}
		ok, err := enc.writeObjectRef(value)
		if err != nil {
			return err
		}
		if ok {
			return nil
		} else {
			enc.refObjects = append(enc.refObjects, value)
			err = writeUTF8(enc.bw, string(*value))
			if err != nil {
				return err
			}
		}
	} else if value, ok := v.(*XMLType); ok {
		_, err := enc.bw.Write([]byte{XmlMarker})
		if err != nil {
			return err
		}
		ok, err := enc.writeObjectRef(value)
		if err != nil {
			return err
		}
		if ok {
			return nil
		} else {
			enc.refObjects = append(enc.refObjects, value)
			err = writeUTF8(enc.bw, string(*value))
			if err != nil {
				return err
			}
		}
	} else if value, ok := v.(*DateType); ok {
		_, err := enc.bw.Write([]byte{DateMarker})
		if err != nil {
			return err
		}
		ok, err := enc.writeObjectRef(value)
		if err != nil {
			return err
		}
		if ok {
			return nil
		} else {
			enc.refObjects = append(enc.refObjects, value)
			f := math.Float64bits(float64(*value))
			binary.BigEndian.PutUint64(u64, f)
			_, err = enc.bw.Write(u64)
			if err != nil {
				return err
			}
		}
	} else if value, ok := v.(*ByteArrayType); ok {
		_, err := enc.bw.Write([]byte{ByteArrayMarker})
		if err != nil {
			return err
		}
		ok, err := enc.writeObjectRef(value)
		if err != nil {
			return err
		}
		if ok {
			return nil
		} else {
			enc.refObjects = append(enc.refObjects, value)
			length := len(*value)
			err = EncodeUInt29(enc.bw, uint32(length<<1))
			if err != nil {
				return err
			}
			enc.bw.Write([]byte(*value))
		}
	}
	return nil
}

func (enc *Encoder) writeString(str StringType) error {
	for i, s := range enc.refStrings {
		if s == str {
			u := uint32(i<<1 | 0x01)
			err := EncodeUInt29(enc.bw, u)
			return err
		}
	}
	err := writeUTF8(enc.bw, string(str))
	if err != nil {
		return err
	}
	enc.refStrings = append(enc.refStrings, str)
	return nil
}

func (enc *Encoder) writeObjectRef(v interface{}) (ok bool, err error) {
	for i, obj := range enc.refObjects {
		if obj == v {
			u := uint32(i<<1 | 0x01)
			err = EncodeUInt29(enc.bw, u)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func writeUTF8(w io.Writer, str string) error {
	length := len(str)
	u := uint32(length << 1)
	err := EncodeUInt29(w, u)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(str))
	if err != nil {
		return err
	}
	return nil
}
