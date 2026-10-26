package identifier

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

const UnicodeDataVersionV1 = "15.0.0"

func unicodeDefaultCasefoldV1(value string) string {
	return norm.NFC.String(cases.Fold().String(norm.NFC.String(value)))
}
