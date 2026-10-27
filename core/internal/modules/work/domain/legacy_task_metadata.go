package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"time"
	"unicode/utf8"
)

const (
	MaxLegacyRecurrenceRuleCharacters = 50
	MaxLegacyFormFilterBytes          = 64_000
)

var ErrLegacyTaskMetadataInvalid = errors.New("work legacy task metadata is invalid")

// LegacyTaskMetadata preserves V1 task fields that are compatibility metadata,
// not the executable V2 calendar recurrence aggregate.
type LegacyTaskMetadata struct {
	IsRecurring    bool
	RecurrenceRule string
	RecurrenceEnd  *time.Time
	FormFilter     json.RawMessage
}

func NormalizeLegacyTaskMetadata(value LegacyTaskMetadata) (LegacyTaskMetadata, error) {
	if !utf8.ValidString(value.RecurrenceRule) ||
		utf8.RuneCountInString(value.RecurrenceRule) > MaxLegacyRecurrenceRuleCharacters {
		return LegacyTaskMetadata{}, ErrLegacyTaskMetadataInvalid
	}
	if value.RecurrenceEnd != nil {
		end := value.RecurrenceEnd.UTC()
		value.RecurrenceEnd = &end
	}
	filter := bytes.TrimSpace(value.FormFilter)
	if len(filter) == 0 || bytes.Equal(filter, []byte("null")) {
		value.FormFilter = nil
		return value, nil
	}
	var document any
	decoder := json.NewDecoder(bytes.NewReader(filter))
	decoder.UseNumber()
	if decoder.Decode(&document) != nil || document == nil {
		return LegacyTaskMetadata{}, ErrLegacyTaskMetadataInvalid
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return LegacyTaskMetadata{}, ErrLegacyTaskMetadataInvalid
	}
	canonical, err := json.Marshal(document)
	if err != nil || len(canonical) > MaxLegacyFormFilterBytes {
		return LegacyTaskMetadata{}, ErrLegacyTaskMetadataInvalid
	}
	value.FormFilter = canonical
	return value, nil
}

func (value LegacyTaskMetadata) Valid() bool {
	normalized, err := NormalizeLegacyTaskMetadata(value)
	return err == nil && bytes.Equal(normalized.FormFilter, value.FormFilter) &&
		normalized.RecurrenceRule == value.RecurrenceRule &&
		timePointersEqual(normalized.RecurrenceEnd, value.RecurrenceEnd)
}

func timePointersEqual(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}
