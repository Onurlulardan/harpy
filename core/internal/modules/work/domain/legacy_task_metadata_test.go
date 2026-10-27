package domain

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeLegacyTaskMetadataCanonicalizesFilterAndUTC(t *testing.T) {
	end := time.Date(2026, 12, 31, 3, 0, 0, 0, time.FixedZone("legacy", 3*60*60))
	value, err := NormalizeLegacyTaskMetadata(LegacyTaskMetadata{
		IsRecurring: true, RecurrenceRule: "FREQ=WEEKLY", RecurrenceEnd: &end,
		FormFilter: []byte(`{"status":"open","priority":1}`)})
	if err != nil || value.RecurrenceEnd == nil ||
		value.RecurrenceEnd.Location() != time.UTC ||
		string(value.FormFilter) != `{"priority":1,"status":"open"}` || !value.Valid() {
		t.Fatalf("metadata normalization drifted: value=%+v err=%v", value, err)
	}
}

func TestNormalizeLegacyTaskMetadataRejectsInvalidOrOversizeValues(t *testing.T) {
	for name, value := range map[string]LegacyTaskMetadata{
		"oversize rule": {RecurrenceRule: strings.Repeat("x",
			MaxLegacyRecurrenceRuleCharacters+1)},
		"invalid filter": {FormFilter: []byte(`{broken`)},
	} {
		if _, err := NormalizeLegacyTaskMetadata(value); err == nil {
			t.Fatalf("%s metadata was accepted", name)
		}
	}
}

func TestNormalizeLegacyTaskMetadataPreservesNullableAndNonObjectJSON(t *testing.T) {
	for _, raw := range []string{"", "null"} {
		value, err := NormalizeLegacyTaskMetadata(LegacyTaskMetadata{FormFilter: []byte(raw)})
		if err != nil || value.FormFilter != nil {
			t.Fatalf("nullable filter %q = %s err=%v", raw, value.FormFilter, err)
		}
	}
	for _, raw := range []string{`[]`, `"record-id"`, `true`, `9007199254740993`} {
		value, err := NormalizeLegacyTaskMetadata(LegacyTaskMetadata{FormFilter: []byte(raw)})
		if err != nil || string(value.FormFilter) != raw {
			t.Fatalf("legacy JSONB filter %q = %s err=%v", raw, value.FormFilter, err)
		}
	}
}

func TestNormalizeLegacyTaskMetadataPreservesExactJSONNumbers(t *testing.T) {
	value, err := NormalizeLegacyTaskMetadata(LegacyTaskMetadata{FormFilter: []byte(
		`{"nested":{"exact":9007199254740993},"decimal":1.2300}`)})
	if err != nil || string(value.FormFilter) !=
		`{"decimal":1.2300,"nested":{"exact":9007199254740993}}` {
		t.Fatalf("legacy filter number changed: filter=%s err=%v", value.FormFilter, err)
	}
}
