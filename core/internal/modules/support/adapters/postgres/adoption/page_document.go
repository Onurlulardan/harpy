package adoption

import (
	"encoding/json"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
	"github.com/google/uuid"
)

type pageDocument struct {
	RunID            uuid.UUID  `json:"run_id"`
	PageNumber       int64      `json:"page_number"`
	CursorBefore     *uuid.UUID `json:"cursor_before"`
	CursorAfter      uuid.UUID  `json:"cursor_after"`
	RowCount         int64      `json:"row_count"`
	SourcePageSHA256 string     `json:"source_page_sha256"`
	TargetPageSHA256 string     `json:"target_page_sha256"`
	Advanced         bool       `json:"advanced"`
}

func decodePage(document []byte) (application.PageReceipt, error) {
	var wire pageDocument
	if len(document) == 0 || json.Unmarshal(document, &wire) != nil ||
		wire.RunID == uuid.Nil || wire.PageNumber < 1 || wire.CursorAfter == uuid.Nil ||
		wire.RowCount < 1 || wire.RowCount > application.PageLimit {
		return application.PageReceipt{}, errDocumentInvalid
	}
	sourceSHA, err := decodeSHA(wire.SourcePageSHA256)
	if err != nil {
		return application.PageReceipt{}, err
	}
	targetSHA, err := decodeSHA(wire.TargetPageSHA256)
	if err != nil {
		return application.PageReceipt{}, err
	}
	return application.PageReceipt{RunID: wire.RunID,
		CursorBefore: valueUUID(wire.CursorBefore), CursorAfter: wire.CursorAfter,
		PageNumber: wire.PageNumber, RowCount: wire.RowCount,
		SourcePageSHA256: sourceSHA, TargetPageSHA256: targetSHA,
		Advanced: wire.Advanced}, nil
}
