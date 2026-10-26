package sqlidentifier

import "strings"

const maximumIdentifierBytes = 63

var dynamicPrefixes = [...]string{"frm_", "fld_", "rel_"}

func validLowerIdentifier(value string) bool {
	if len(value) < 1 || len(value) > maximumIdentifierBytes ||
		value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

func hasDynamicPrefix(value string) bool {
	for _, prefix := range dynamicPrefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
