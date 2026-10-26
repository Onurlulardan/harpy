package lifecycle

import "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"

func validStableID(raw string) bool {
	value, err := packagecontract.ParseStableID(raw)
	return err == nil && value.String() == raw
}
