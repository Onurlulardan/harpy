package recordkernel

func expressionType(raw string) (TypeID, bool) {
	typeID := TypeID(raw)
	switch typeID {
	case TypeText, TypeInteger, TypeDecimal,
		TypeBoolean, TypeDate, TypeTimestamp,
		TypeTime:
		return typeID, true
	default:
		return "", false
	}
}

func numericType(typeID TypeID) bool {
	return typeID == TypeInteger || typeID == TypeDecimal
}

func orderedType(typeID TypeID) bool {
	switch typeID {
	case TypeText, TypeInteger, TypeDecimal,
		TypeDate, TypeTimestamp, TypeTime:
		return true
	default:
		return false
	}
}
