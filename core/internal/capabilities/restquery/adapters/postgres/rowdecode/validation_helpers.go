package rowdecode

import (
	"strings"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/platform/textprofile"
)

func validText(value string) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, '\x00') &&
		textprofile.ExactNFC(value) == value
}

func zeroLastValue() cursor.LastValue { return cursor.LastValue{} }
