package amf0

import (
	"bytes"
	"testing"
)

func TestWriteUTF8(t *testing.T) {
	buf := new(bytes.Buffer)
	err := writeUTF8(buf, "foo")
	if err != nil {
		t.Errorf("test for %s error: %s", "foo", err)
	} else {
		expect := []byte{0x00, 0x03, 'f', 'o', 'o'}
		got := buf.Bytes()
		if !bytes.Equal(expect, got) {
			t.Errorf("expect %x got %x", expect, got)
		}
	}
	buf = new(bytes.Buffer)
	err = writeUTF8(buf, "你好")
	if err != nil {
		t.Errorf("test for %s error: %s", "你好", err)
	} else {
		expect := []byte{0x00, 0x06, 0xe4, 0xbd, 0xa0, 0xe5, 0xa5, 0xbd}
		got := buf.Bytes()
		if !bytes.Equal(expect, got) {
			t.Errorf("expect %x got %x", expect, got)
		}
	}
}

func TestWriteUTF8Long(t *testing.T) {
	buf := new(bytes.Buffer)
	err := writeUTF8Long(buf, "foo")
	if err != nil {
		t.Errorf("test for %s error: %s", "foo", err)
	} else {
		expect := []byte{0x00, 0x00, 0x00, 0x03, 'f', 'o', 'o'}
		got := buf.Bytes()
		if !bytes.Equal(expect, got) {
			t.Errorf("expect %x got %x", expect, got)
		}
	}
	buf = new(bytes.Buffer)
	err = writeUTF8Long(buf, "你好")
	if err != nil {
		t.Errorf("test for %s error: %s", "你好", err)
	} else {
		expect := []byte{0x00, 0x00, 0x00, 0x06, 0xe4, 0xbd, 0xa0, 0xe5, 0xa5, 0xbd}
		got := buf.Bytes()
		if !bytes.Equal(expect, got) {
			t.Errorf("expect %x got %x", expect, got)
		}
	}
}

func TestEncodeNumber(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	num := NumberType(1.2)
	err := enc.Encode(num)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x00, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Fatalf("expect %x got %x", expect, got)
	}
}

func TestEncodeBoolean(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	b := BooleanType(true)
	err := enc.Encode(b)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x01, 0x01}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeString(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	s := StringType("foo")
	err := enc.Encode(s)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x02, 0x00, 0x03, 0x66, 0x6f, 0x6f}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeLongString(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	s := LongStringType("foo")
	err := enc.Encode(s)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x0c, 0x00, 0x00, 0x00, 0x03, 0x66, 0x6f, 0x6f}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeObject(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	var obj ObjectType
	obj = make(ObjectType)
	obj["foo"] = StringType("bar")
	err := enc.Encode(&obj)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x03, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x02, 0x00, 0x03, 0x62, 0x61, 0x72, 0x00, 0x00, 0x09}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeObjectReference(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	var obj ObjectType
	obj = make(ObjectType)
	obj["foo"] = &obj
	err := enc.Encode(&obj)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x03, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x07, 0x00, 0x00, 0x00, 0x00, 0x09}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeEcmaArray(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	var obj EcmaArrayType
	obj = make(EcmaArrayType)
	obj["foo"] = StringType("bar")
	err := enc.Encode(&obj)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x02, 0x00, 0x03, 0x62, 0x61, 0x72, 0x00, 0x00, 0x09}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeStrictArray(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	var obj StrictArrayType
	obj = make(StrictArrayType, 3)
	obj[0] = NumberType(5)
	obj[1] = StringType("foo")
	obj[2] = NullType{}
	err := enc.Encode(&obj)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x0a, 0x00, 0x00, 0x00, 0x03, 0x00, 0x40, 0x14, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x05}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeDate(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	var obj DateType
	obj.Date = 5
	err := enc.Encode(obj)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x0b, 0x40, 0x14, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeNull(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	nt := NullType{}
	err := enc.Encode(nt)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x05}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeUndefined(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	nt := UndefinedType{}
	err := enc.Encode(nt)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x06}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeXmlDocument(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	xml := XmlDocumentType("xml")
	err := enc.Encode(xml)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x0f, 0x00, 0x00, 0x00, 0x03, 0x78, 0x6d, 0x6c}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeTypedObject(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	obj := new(TypedObjectType)
	obj.ClassName = StringType("class")
	obj.Object = make(_Object)
	obj.Object["self"] = obj
	err := enc.Encode(obj)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect := []byte{0x10, 0x00, 0x05, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x00, 0x04, 0x73, 0x65, 0x6c, 0x66, 0x07, 0x00, 0x00, 0x00,
		0x00, 0x09}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}
