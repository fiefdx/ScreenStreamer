package amf0

import ()

type NullType struct {
}

type UndefinedType struct {
}

type UnsupportedType struct {
}

type NumberType float64
type BooleanType bool
type StringType string
type LongStringType string
type XmlDocumentType LongStringType
type _Object map[StringType]interface{}
type ObjectType _Object
type EcmaArrayType _Object
type StrictArrayType []interface{}

type DateType struct {
	TimeZone int16
	Date     float64
}

type TypedObjectType struct {
	ClassName StringType
	Object    _Object
}
