package amf3

import (
	"bufio"
	"encoding/binary"
	"errors"
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
	r          io.Reader
	refStrings []StringType  // Strings
	refObjects []interface{} // Object, Array, XML, XMLDocument, ByteArray, Date and instances of user defined Classes
	refTraits  []*Trait      // Objects and instances of user defined Classes have trait information
}

func NewDecoder(r io.Reader) *Decoder {
	if _, ok := r.(*bufio.Reader); ok {
		return &Decoder{r: ConsistReader{rd: r}}
	}
	return &Decoder{r: ConsistReader{rd: bufio.NewReader(r)}}
}

func (dec *Decoder) Decode() (interface{}, error) {
	v, err := dec.decodeValue()
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (dec *Decoder) decodeValue() (interface{}, error) {
	u8 := make([]byte, 1)
	_, err := dec.r.Read(u8)
	if err != nil {
		return nil, err
	}
	switch u8[0] {
	case UndefinedMarker:
		return UndefinedType{}, nil
	case NullMarker:
		return NullType{}, nil
	case FalseMarker:
		return FalseType{}, nil
	case TrueMarker:
		return TrueType{}, nil
	case IntegerMarker:
		i, err := DecodeUInt29(dec.r)
		if err != nil {
			return nil, err
		}
		return IntegerType(i), nil
	case DoubleMarker:
		f, err := dec.readFloat()
		if err != nil {
			return nil, err
		}
		return DoubleType(f), nil
	case StringMarker:
		str, err := dec.readString()
		if err != nil {
			return nil, err
		}
		return StringType(str), nil
	case XmlDocMarker:
		ref, i, err := dec.readRefInt()
		if err != nil {
			return nil, err
		}
		xmldoc := new(XMLDocumentType)
		if ref {
			obj, err := dec.getRefObject(i)
			if err != nil {
				return nil, err
			}
			var ok bool
			if xmldoc, ok = obj.(*XMLDocumentType); !ok {
				return nil, errors.New("wrong ref type")
			}
		} else {
			strBytes := make([]byte, i)
			_, err = dec.r.Read(strBytes)
			if err != nil {
				return nil, err
			}
			*xmldoc = XMLDocumentType(strBytes)
			dec.refObjects = append(dec.refObjects, xmldoc)
		}
		return xmldoc, nil
	case DateMarker:
		ref, i, err := dec.readRefInt()
		if err != nil {
			return nil, err
		}
		date := new(DateType)
		if ref {
			obj, err := dec.getRefObject(i)
			if err != nil {
				return nil, err
			}
			var ok bool
			if date, ok = obj.(*DateType); !ok {
				return nil, errors.New("wrong ref type")
			}
		} else {
			f, err := dec.readFloat()
			if err != nil {
				return nil, err
			}
			*date = DateType(f)
			dec.refObjects = append(dec.refObjects, date)
		}
		return date, nil
	case ArrayMarker:
		ref, i, err := dec.readRefInt()
		if ref {
			obj, err := dec.getRefObject(i)
			if err != nil {
				return nil, err
			}
			if _, ok := obj.(*ArrayType); !ok {
				return nil, errors.New("wrong ref type")
			}
			return obj, nil
		} else {
			denseCount := i
			array := new(ArrayType)
			dec.refObjects = append(dec.refObjects, array)
			for {
				s, err := dec.readString()
				if err != nil {
					return nil, err
				}
				if s == "" {
					break
				}
				array.Associative[s], err = dec.decodeValue()
				if err != nil {
					return nil, err
				}
			}
			array.Dense = make([]interface{}, denseCount)
			for k := 0; k < int(denseCount); k++ {
				array.Dense[k], err = dec.decodeValue()
				if err != nil {
					return nil, err
				}
			}
			return array, nil
		}
	case XmlMarker:
		ref, i, err := dec.readRefInt()
		if err != nil {
			return nil, err
		}
		xml := new(XMLType)
		if ref {
			obj, err := dec.getRefObject(i)
			if err != nil {
				return nil, err
			}
			var ok bool
			if xml, ok = obj.(*XMLType); !ok {
				return nil, errors.New("wrong ref type")
			}
		} else {
			strBytes := make([]byte, i)
			_, err = dec.r.Read(strBytes)
			if err != nil {
				return nil, err
			}
			*xml = XMLType(strBytes)
			dec.refObjects = append(dec.refObjects, xml)
		}
		return xml, nil
	case ByteArrayMarker:
		ref, i, err := dec.readRefInt()
		if err != nil {
			return nil, err
		}
		if ref {
			obj, err := dec.getRefObject(i)
			if err != nil {
				return nil, err
			}
			if _, ok := obj.(*ByteArrayType); !ok {
				return nil, errors.New("wrong ref type")
			}
			return obj, nil
		} else {
			byteArray := make([]byte, i)
			_, err = dec.r.Read(byteArray)
			if err != nil {
				return nil, err
			}
			pbyteArray := new(ByteArrayType)
			*pbyteArray = ByteArrayType(byteArray)
			dec.refObjects = append(dec.refObjects, pbyteArray)
			return pbyteArray, nil
		}
	case ObjectMarker:
		ref, i, err := dec.readRefInt()
		if err != nil {
			return nil, err
		}

		if ref {
			obj, err := dec.getRefObject(i)
			if err != nil {
				return nil, err
			}
			if _, ok := obj.(*ObjectType); !ok {
				return nil, errors.New("wrong ref type")
			}
			return obj, nil
		} else {
			obj := new(ObjectType)
			dec.refObjects = append(dec.refObjects, obj)
			var trait *Trait
			if i&0x01 == 0 {
				trait, err = dec.getRefTrait(i >> 1)
				if err != nil {
					return nil, err
				}
			} else {
				trait = new(Trait)
				if i&0x02 == 0 {
					trait.IsDynamic = i&0x04 == 1
					attrsCount := int(i >> 3)
					trait = new(Trait)
					trait.ClassName, err = dec.readString()
					if err != nil {
						return nil, err
					}
					trait.Attrs = make([]StringType, attrsCount)
					for k := 0; k < attrsCount; k++ {
						trait.Attrs[k], err = dec.readString()
						if err != nil {
							return nil, err
						}
					}
					dec.refTraits = append(dec.refTraits, trait)
				} else {
					return nil, errors.New("traits-ext not support")
				}
				obj.Trait = trait
				obj.Static = make([]interface{}, len(trait.Attrs))
				for k := 0; k < len(trait.Attrs); k++ {
					obj.Static[k], err = dec.decodeValue()
					if err != nil {
						return nil, err
					}
				}
				obj.Dynamic = make(map[StringType]interface{})
				if trait.IsDynamic {
					for {
						name, err := dec.readString()
						if err != nil {
							return nil, err
						}
						if name == "" {
							break
						}
						obj.Dynamic[name], err = dec.decodeValue()
						if err != nil {
							return nil, err
						}
					}
				}
			}
			return obj, nil
		}
	}
	return nil, nil
}

func (dec *Decoder) readRefInt() (ref bool, i uint32, err error) {
	u29, err := DecodeUInt29(dec.r)
	if err != nil {
		return
	}
	ref = u29&0x01 == 0
	i = u29 >> 1
	return
}

func (dec *Decoder) readFloat() (float64, error) {
	u64 := make([]byte, 8)
	_, err := dec.r.Read(u64)
	if err != nil {
		return 0, err
	}
	u64n := binary.BigEndian.Uint64(u64)
	f := math.Float64frombits(u64n)
	if err != nil {
		return 0, err
	}
	return f, nil
}

func (dec *Decoder) readString() (StringType, error) {
	ref, i, err := dec.readRefInt()
	if err != nil {
		return "", err
	}
	var str StringType
	if ref {
		str, err = dec.getRefString(i)
		if err != nil {
			return "", err
		}
	} else {
		strBytes := make([]byte, i)
		_, err = dec.r.Read(strBytes)
		if err != nil {
			return "", err
		}
		str = StringType(strBytes)
		if str != "" {
			dec.refStrings = append(dec.refStrings, str)
		}
	}
	return str, nil
}

func (dec *Decoder) getRefString(i uint32) (StringType, error) {
	if int(i) >= len(dec.refStrings) {
		return "", errors.New("refStrings index outbound")
	}
	return dec.refStrings[i], nil
}

func (dec *Decoder) getRefObject(i uint32) (interface{}, error) {
	if int(i) >= len(dec.refObjects) {
		return nil, errors.New("refObjects index outbound")
	}
	return dec.refObjects[i], nil
}

func (dec *Decoder) getRefTrait(i uint32) (*Trait, error) {
	if int(i) >= len(dec.refTraits) {
		return nil, errors.New("refTraits index outbound")
	}
	return dec.refTraits[i], nil
}
