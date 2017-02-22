package amf0

import (
	"bytes"
	"testing"
)

func TestReadUTF8(t *testing.T) {
	buf := bytes.NewReader([]byte{0x00, 0x03, 'f', 'o', 'o'})
	expect := StringType("foo")
	got, err := readUTF8(buf)
	if err != nil {
		t.Errorf("test for %s error: %s", "foo", err)
	} else {
		if expect != got {
			t.Errorf("expect %v got %v", expect, got)
		}
	}
	buf = bytes.NewReader([]byte{0x00, 0x06, 0xe4, 0xbd, 0xa0, 0xe5, 0xa5, 0xbd})
	expect = StringType("你好")
	got, err = readUTF8(buf)
	if err != nil {
		t.Errorf("test for %s error: %s", "你好", err)
	} else {
		if expect != got {
			t.Errorf("expect %v got %v", expect, got)
		}
	}
}

func TestReadUTF8Long(t *testing.T) {
	buf := bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x03, 'f', 'o', 'o'})
	expect := LongStringType("foo")
	got, err := readUTF8Long(buf)
	if err != nil {
		t.Errorf("test for %s error: %s", "foo", err)
	} else {
		if expect != got {
			t.Errorf("expect %v got %v", expect, got)
		}
	}
	buf = bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x06, 0xe4, 0xbd, 0xa0, 0xe5, 0xa5, 0xbd})
	expect = LongStringType("你好")
	got, err = readUTF8Long(buf)
	if err != nil {
		t.Errorf("test for %s error: %s", "你好", err)
	} else {
		if expect != got {
			t.Errorf("expect %v got %v", expect, got)
		}
	}
}

func TestDecodeNumber(t *testing.T) {
	buf := bytes.NewReader([]byte{0x00, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33})
	expect := NumberType(1.2)
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if expect != got {
		t.Fatalf("expect %v got %v", expect, got)
	}
}

func TestDecodeBoolean(t *testing.T) {
	buf := bytes.NewReader([]byte{0x01, 0x01})
	expect := BooleanType(true)
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if expect != got {
		t.Fatalf("expect %v got %v", expect, got)
	}
}

func TestDecodeString(t *testing.T) {
	buf := bytes.NewReader([]byte{0x02, 0x00, 0x03, 0x66, 0x6f, 0x6f})
	expect := StringType("foo")
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if expect != got {
		t.Fatalf("expect %v got %v", expect, got)
	}
}

func TestDecodeLongString(t *testing.T) {
	buf := bytes.NewReader([]byte{0x0c, 0x00, 0x00, 0x00, 0x03, 0x66, 0x6f, 0x6f})
	expect := LongStringType("foo")
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if expect != got {
		t.Fatalf("expect %v got %v", expect, got)
	}
}

func TestDecodeObject(t *testing.T) {
	buf := bytes.NewReader([]byte{0x03, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x02, 0x00, 0x03, 0x62, 0x61, 0x72, 0x00, 0x00, 0x09})
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	obj, ok := got.(*ObjectType)
	if !ok {
		t.Fatalf("type incorrect")
	}
	if (*obj)["foo"] != StringType("bar") {
		t.Fatalf("decode incorrect")
	}
}

func TestDecodeObjectReference(t *testing.T) {
	buf := bytes.NewReader([]byte{0x03, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x07, 0x00, 0x00, 0x00, 0x00, 0x09})
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	obj, ok := got.(*ObjectType)
	if !ok {
		t.Fatalf("type incorrect")
	}
	if (*obj)["foo"] != obj {
		t.Fatalf("decode incorrect")
	}
}

func TestDecodeObjectEcmaArray(t *testing.T) {
	buf := bytes.NewReader([]byte{0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x02, 0x00, 0x03, 0x62, 0x61, 0x72, 0x00, 0x00, 0x09})
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	obj, ok := got.(*EcmaArrayType)
	if !ok {
		t.Fatalf("type incorrect")
	}
	if (*obj)["foo"] != StringType("bar") {
		t.Fatalf("decode incorrect")
	}
}

func TestDecodeStrictArray(t *testing.T) {
	buf := bytes.NewReader([]byte{0x0a, 0x00, 0x00, 0x00, 0x03, 0x00, 0x40, 0x14, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x05})
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	obj, ok := got.(*StrictArrayType)
	if !ok {
		t.Fatalf("type incorrect")
	}
	if (*obj)[0] != NumberType(5) {
		t.Fatalf("decode incorrect")
	}
	if (*obj)[1] != StringType("foo") {
		t.Fatalf("decode incorrect")
	}
	if (*obj)[2] != (NullType{}) {
		t.Fatalf("decode incorrect")
	}
}

func TestDecodeDate(t *testing.T) {
	buf := bytes.NewReader([]byte{0x0b, 0x40, 0x14, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	expect := DateType{Date: 5}
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if expect != got {
		t.Fatalf("expect %v got %v", expect, got)
	}
}

func TestDecodeNull(t *testing.T) {
	buf := bytes.NewReader([]byte{0x05})
	expect := NullType{}
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if expect != got {
		t.Fatalf("expect %v got %v", expect, got)
	}
}

func TestDecodeUndefined(t *testing.T) {
	buf := bytes.NewReader([]byte{0x06})
	expect := UndefinedType{}
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if expect != got {
		t.Fatalf("expect %v got %v", expect, got)
	}
}

func TestDecodeXmlDocument(t *testing.T) {
	buf := bytes.NewReader([]byte{0x0f, 0x00, 0x00, 0x00, 0x03, 0x78, 0x6d, 0x6c})
	expect := XmlDocumentType("xml")
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if expect != got {
		t.Fatalf("expect %v got %v", expect, got)
	}
}

func TestDecodeTypedObject(t *testing.T) {
	buf := bytes.NewReader([]byte{0x10, 0x00, 0x05, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x00, 0x04, 0x73, 0x65, 0x6c, 0x66, 0x07, 0x00, 0x00, 0x00,
		0x00, 0x09})
	dec := NewDecoder(buf)
	got, err := dec.Decode()
	if err != nil {
		t.Fatalf("%s", err)
	}
	obj, ok := got.(*TypedObjectType)
	if !ok {
		t.Fatalf("type incorrect")
	}
	if (*obj).ClassName != "class" {
		t.Fatalf("decode error %s", (*obj).ClassName)
	}
	if (*obj).Object["self"] != obj {
		t.Fatalf("decode error")
	}
}
