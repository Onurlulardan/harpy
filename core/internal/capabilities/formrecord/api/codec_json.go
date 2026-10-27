package recordkernel

import (
	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func decodeJSON(typeID TypeID, raw []byte) (Value, error) {
	canonical, err := canonicaljsonv1.Canonicalize(raw, MaximumFieldBytes)
	if err != nil {
		return Value{}, ErrInputInvalid
	}
	return typedValue(typeID, canonical), nil
}
