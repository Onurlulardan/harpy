package adoption

import (
	"crypto/sha256"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

type LegacyRecord struct {
	SourceID                         uuid.UUID
	EnvironmentID, TenantID, ActorID uuid.UUID
	Method, Path                     string
	StatusCode                       int
	DurationMicros                   int64
	OccurredAt                       time.Time
}

type SourcePage struct {
	Records     []LegacyRecord
	CursorAfter uuid.UUID
	SHA256      [32]byte
}

func NewSourcePage(records []LegacyRecord) (SourcePage, error) {
	page := SourcePage{Records: append([]LegacyRecord(nil), records...)}
	if len(records) > 0 {
		page.CursorAfter = records[len(records)-1].SourceID
	}
	document := make([]byte, 0, len(records)*32)
	for _, record := range records {
		digest := sourceDigest(record)
		document = append(document, digest[:]...)
	}
	page.SHA256 = sha256.Sum256(document)
	if page.Validate(100) != nil {
		return SourcePage{}, domain.ErrInvalid
	}
	return page, nil
}

func (page SourcePage) Validate(limit int) error {
	if len(page.Records) < 1 || len(page.Records) > limit || limit < 1 || limit > 100 ||
		page.CursorAfter == uuid.Nil || page.SHA256 == [32]byte{} ||
		page.Records[len(page.Records)-1].SourceID != page.CursorAfter {
		return domain.ErrInvalid
	}
	previous := uuid.Nil
	for _, record := range page.Records {
		if record.SourceID == uuid.Nil || record.EnvironmentID == uuid.Nil ||
			record.TenantID == uuid.Nil || record.ActorID == uuid.Nil ||
			record.Method == "" || record.Path == "" || record.StatusCode < 100 ||
			record.StatusCode > 599 || record.DurationMicros < 0 ||
			record.OccurredAt.IsZero() || record.OccurredAt.Nanosecond()%1000 != 0 ||
			record.OccurredAt.UnixMilli() < 0 || record.OccurredAt.UnixMilli() >= 1<<48 ||
			previous != uuid.Nil && string(record.SourceID[:]) <= string(previous[:]) {
			return domain.ErrInvalid
		}
		previous = record.SourceID
	}
	return nil
}
