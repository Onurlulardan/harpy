package documents

import "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"

func parseDocumentID(value string) (foundation.ID, error) {
	return foundation.ParseID(value)
}
