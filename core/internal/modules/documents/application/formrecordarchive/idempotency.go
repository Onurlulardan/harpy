package formrecordarchive

import (
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func formRecordIdempotency(key string, definitionID, recordID domain.ID) string {
	digest := foundation.Digest([]byte("documents.form_record_archive.v1"),
		[]byte(key), definitionID[:], recordID[:])
	return "forms-record-archive-" + hex.EncodeToString(digest[:])
}
