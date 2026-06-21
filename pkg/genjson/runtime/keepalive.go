package runtime

// Keep the exported helper surface referenced inside the package so static analysis
// does not report it as unused. These symbols are consumed by generated code.
var _ = [...]any{
	IsZero,
	DecodeObject,
	KeyEq,
	IsNull,
	DecodeString,
	DecodeText,
	DecodeArray,
	DecodeMapObject,
	AppendCommaIfNeeded,
	AppendArrayStart,
	AppendArrayEnd,
	AppendObjectStart,
	AppendObjectEnd,
	DecodeBool,
	DecodeInt,
	DecodeUint,
	DecodeFloat,
	AppendFieldName,
	AppendFieldToken,
	AppendFloat64,
}
