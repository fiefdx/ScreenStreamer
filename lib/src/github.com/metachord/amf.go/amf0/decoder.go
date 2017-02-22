package amf0

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

type ConsistReader struct {
	rd io.Reader
}

func (cr ConsistReader) Read(p []byte) (n int, err error) {
	pos := 0
	for {
		n, err := cr.rd.Read(p[pos:])
		if err != nil {
			return n, err
		}
		pos += n
		if pos == len(p) {
			break
		}
	}
	return pos, nil
}

type Decoder struct {
	r       ConsistReader
	refObjs []interface{}
}

// should use io.LimitedReader
func NewDecoder(r io.Reader) *Decoder {
	if _, ok := r.(*bufio.Reader); ok {
		return &Decoder{r: ConsistReader{rd: r}}
	}
	return &Decoder{r: ConsistReader{rd: bufio.NewReader(r)}}
}

func (dec *Decoder) Decode() (v interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			v = nil
			err = errors.New(r.(string))
		}
	}()
	v, err = dec.decodeValue()
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (dec *Decoder) decodeValue() (interface{}, error) {
	u8 := make([]byte, 1)
	u16 := make([]byte, 2)
	u32 := make([]byte, 4)
	u64 := make([]byte, 8)
	_, err := dec.r.Read(u8)
	if err != nil {
		return nil, err
	}
	marker := u8[0]
	switch marker {
	case NumberMarker:
		_, err := dec.r.Read(u64)
		if err != nil {
			return nil, err
		}
		u64n := binary.BigEndian.Uint64(u64)
		number := math.Float64frombits(u64n)
		return NumberType(number), nil
	case BooleanMarker:
		_, err := dec.r.Read(u8)
		if err != nil {
			return nil, err
		}
		return BooleanType(u8[0] != 0), nil
	case StringMarker:
		stringBytes, err := readUTF8(dec.r)
		if err != nil {
			return nil, err
		}
		return StringType(stringBytes), nil
	case ObjectMarker:
		object := new(ObjectType)
		dec.refObjs = append(dec.refObjs, object)
		obj, err := dec.readObject()
		if err != nil {
			return nil, err
		}
		*object = ObjectType(obj)
		return object, nil
	case MovieclipMarker:
		return nil, errors.New("Movieclip Type not supported")
	case NullMarker:
		return NullType{}, nil
	case UndefinedMarker:
		return UndefinedType{}, nil
	case ReferenceMarker:
		_, err = dec.r.Read(u16)
		if err != nil {
			return nil, err
		}
		refid := binary.BigEndian.Uint16(u16)
		if int(refid) >= len(dec.refObjs) {
			return nil, errors.New("reference error")
		}
		return dec.refObjs[refid], nil
	case EcmaArrayMarker:
		_, err := dec.r.Read(u32)
		if err != nil {
			return nil, err
		}
		object := new(EcmaArrayType)
		dec.refObjs = append(dec.refObjs, object)
		associativeCount := binary.BigEndian.Uint32(u32)
		obj, err := dec.readObject()
		if err != nil {
			return nil, err
		}
		*object = EcmaArrayType(obj)
		if uint32(len(*object)) != associativeCount {
			return nil, errors.New("EcmaArray count error")
		}
		return object, nil
	case StrictArrayMarker:
		_, err := dec.r.Read(u32)
		if err != nil {
			return nil, err
		}
		object := new(StrictArrayType)
		dec.refObjs = append(dec.refObjs, object)
		arrayCount := binary.BigEndian.Uint32(u32)
		array := make(StrictArrayType, arrayCount)
		for i := 0; i < int(arrayCount); i++ {
			array[i], err = dec.decodeValue()
			if err != nil {
				return nil, err
			}
		}
		*object = array
		return object, nil
	case DateMarker:
		_, err := dec.r.Read(u64)
		if err != nil {
			return nil, err
		}
		u64n := binary.BigEndian.Uint64(u64)
		date := math.Float64frombits(u64n)
		_, err = dec.r.Read(u16)
		if err != nil {
			return nil, err
		}
		return DateType{Date: date}, nil
	case LongStringMarker:
		stringBytes, err := readUTF8Long(dec.r)
		if err != nil {
			return nil, err
		}
		return LongStringType(stringBytes), nil
	case UnsupportedMarker:
		return UnsupportedType{}, nil
	case RecordsetMarker:
		return nil, errors.New("RecordSet Type not supported")
	case XmlDocumentMarker:
		stringBytes, err := readUTF8Long(dec.r)
		if err != nil {
			return nil, err
		}
		return XmlDocumentType(stringBytes), nil
	case TypedObjectMarker:
		object := new(TypedObjectType)
		dec.refObjs = append(dec.refObjs, object)
		classNameBytes, err := readUTF8(dec.r)
		if err != nil {
			return nil, err
		}
		obj, err := dec.readObject()
		if err != nil {
			return nil, err
		}
		*object = TypedObjectType{ClassName: StringType(classNameBytes), Object: _Object(obj)}
		return object, nil
	}
	panic(fmt.Sprintf("not reach: %#v", marker))
	return nil, nil
}

func (dec *Decoder) readObject() (_Object, error) {
	u8 := make([]byte, 1)
	v := make(map[StringType]interface{})
	for {
		name, err := readUTF8(dec.r)
		if err != nil {
			return nil, err
		}
		if name == "" {
			_, err := dec.r.Read(u8)
			if err != nil {
				return nil, err
			}
			if u8[0] == ObjectEndMarker {
				break
			} else {
				return nil, errors.New("expect ObjectEndMarker here")
			}
		}
		value, err := dec.decodeValue()
		if err != nil {
			return nil, err
		}
		if _, ok := v[name]; ok {
			return nil, errors.New("object-property exists")
		}
		v[name] = value
	}
	return v, nil
}

func readUTF8(r io.Reader) (StringType, error) {
	u16 := make([]byte, 2)
	_, err := r.Read(u16)
	if err != nil {
		return "", err
	}
	stringLength := binary.BigEndian.Uint16(u16)
	if stringLength == 0 {
		return "", nil
	}
	stringBytes := make([]byte, stringLength)
	_, err = r.Read(stringBytes)
	if err != nil {
		return "", err
	}
	return StringType(stringBytes), nil
}

func readUTF8Long(r io.Reader) (LongStringType, error) {
	u32 := make([]byte, 4)
	_, err := r.Read(u32)
	if err != nil {
		return "", err
	}
	stringLength := binary.BigEndian.Uint32(u32)
	if stringLength == 0 {
		return "", nil
	}
	stringBytes := make([]byte, stringLength)
	_, err = r.Read(stringBytes)
	if err != nil {
		return "", err
	}
	return LongStringType(stringBytes), nil
}
