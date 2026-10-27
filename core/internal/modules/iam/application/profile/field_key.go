package profile

func validFieldKey(value string) bool {
	if len(value) < 1 || len(value) > 63 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	previousUnderscore := false
	for _, character := range value {
		valid := character >= 'a' && character <= 'z' ||
			character >= '0' && character <= '9' || character == '_'
		if !valid || character == '_' && previousUnderscore {
			return false
		}
		previousUnderscore = character == '_'
	}
	return !previousUnderscore
}
