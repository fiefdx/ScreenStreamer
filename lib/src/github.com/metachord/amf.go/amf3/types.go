package amf3

type UndefinedType struct {
}

type NullType struct {
}

type FalseType struct {
}

type TrueType struct {
}

type IntegerType uint32
type DoubleType float64
type StringType string
type XMLDocumentType string
type DateType float64
type ArrayType struct {
	Associative map[StringType]interface{}
	Dense       []interface{}
}

type Trait struct {
	ClassName StringType
	IsDynamic bool
	Attrs     []StringType
}

type ObjectType struct {
	Trait   *Trait
	Static  []interface{}
	Dynamic map[StringType]interface{}
}

type XMLType string
type ByteArrayType []byte
