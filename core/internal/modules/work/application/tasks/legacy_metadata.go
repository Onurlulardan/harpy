package tasks

import (
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func normalizeLegacyMetadata(value *domain.LegacyTaskMetadata) error {
	if value == nil {
		return nil
	}
	normalized, err := domain.NormalizeLegacyTaskMetadata(*value)
	if err != nil {
		return ports.ErrInvalid
	}
	*value = normalized
	return nil
}

func legacyMetadataFingerprint(replace bool, value *domain.LegacyTaskMetadata) string {
	if !replace {
		return "preserve"
	}
	if value == nil {
		return "invalid"
	}
	return strconv.FormatBool(value.IsRecurring) + "\x00" + value.RecurrenceRule +
		"\x00" + timeString(value.RecurrenceEnd) + "\x00" + string(value.FormFilter)
}
