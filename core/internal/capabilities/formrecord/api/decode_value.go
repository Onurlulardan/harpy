package recordkernel

import (
	"bytes"
	"fmt"
)

func DecodeValue(field Field, raw []byte) (Value, error) {
	if bytes.Equal(raw, []byte("null")) {
		if !field.Nullable {
			return Value{}, ErrInputInvalid
		}
		return nullValue(field.TypeID), nil
	}
	var value Value
	var err error
	switch field.TypeID {
	case TypeText:
		value, err = decodeText(field.TypeID, raw)
	case TypeInteger:
		value, err = decodeInteger(field.TypeID, raw)
	case TypeDecimal:
		value, err = decodeDecimal(field, raw)
	case TypeBoolean:
		value, err = decodeBoolean(field.TypeID, raw)
	case TypeDate, TypeTimestamp, TypeTime:
		value, err = decodeTemporal(field.TypeID, raw)
	case TypeUUID, TypeRelation, TypeDocumentRef:
		value, err = decodeUUID(field.TypeID, raw)
	case TypeEnum:
		value, err = decodeEnum(field, raw)
	case TypeTextList, TypeUUIDList,
		TypeDocumentRefList, TypeEnumList:
		value, err = decodeList(field, raw)
	case TypeJSON:
		value, err = decodeJSON(field.TypeID, raw)
	default:
		err = ErrInputInvalid
	}
	if err != nil || len(value.canonical) > MaximumFieldBytes {
		return Value{}, fmt.Errorf("%w: typed field", ErrInputInvalid)
	}
	return value, nil
}

func valuesEqual(left, right Value) bool {
	return left.typeID == right.typeID && left.null == right.null &&
		bytes.Equal(left.canonical, right.canonical)
}
