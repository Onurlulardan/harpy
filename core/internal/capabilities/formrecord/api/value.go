package recordkernel

import (
	"math/big"
)

type Value struct {
	typeID    TypeID
	null      bool
	canonical []byte
	decimal   *big.Rat
}

func (value Value) TypeID() TypeID { return value.typeID }
func (value Value) IsNull() bool   { return value.null }

func (value Value) CanonicalJSON() []byte {
	return append([]byte(nil), value.canonical...)
}

func (value Value) clone() Value {
	value.canonical = append([]byte(nil), value.canonical...)
	if value.decimal != nil {
		value.decimal = new(big.Rat).Set(value.decimal)
	}
	return value
}

func decimalValue(value *big.Rat) Value {
	return Value{typeID: TypeDecimal, decimal: new(big.Rat).Set(value)}
}

func nullValue(typeID TypeID) Value {
	return Value{typeID: typeID, null: true, canonical: []byte("null")}
}

func typedValue(typeID TypeID, canonical []byte) Value {
	return Value{typeID: typeID, canonical: append([]byte(nil), canonical...)}
}
