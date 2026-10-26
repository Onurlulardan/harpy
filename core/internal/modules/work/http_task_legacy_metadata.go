package work

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func createLegacyTaskMetadata(
	wire writeWire, compatibility bool,
) (*domain.LegacyTaskMetadata, error) {
	if !compatibility {
		return nil, nil
	}
	value := domain.LegacyTaskMetadata{RecurrenceRule: wire.RecurrenceRule}
	if wire.IsRecurring != nil {
		value.IsRecurring = *wire.IsRecurring
	}
	value.RecurrenceEnd = legacyRecurrenceEnd(wire.RecurrenceEnd, nil)
	var err error
	normalized, err := domain.NormalizeLegacyTaskMetadata(value)
	if err != nil {
		return nil, ports.ErrInvalid
	}
	return &normalized, nil
}

func patchLegacyTaskMetadata(
	ctx context.Context,
	store ports.TaskStore,
	taskID domain.ID,
	compatibility bool,
	wire writeWire,
) (bool, domain.LegacyTaskMetadata, error) {
	if !compatibility || !legacyMetadataPatchPresent(wire) {
		return false, domain.LegacyTaskMetadata{}, nil
	}
	value, err := store.LoadLegacyTaskMetadataForUpdate(ctx, taskID)
	if err != nil {
		return false, domain.LegacyTaskMetadata{}, err
	}
	if presentCompatibilityField(wire, "is_recurring") {
		if wire.IsRecurring == nil {
			return false, domain.LegacyTaskMetadata{}, ports.ErrInvalid
		}
		value.IsRecurring = *wire.IsRecurring
	}
	if presentCompatibilityField(wire, "recurrence_rule") {
		value.RecurrenceRule = wire.RecurrenceRule
	}
	if presentCompatibilityField(wire, "recurrence_end") {
		value.RecurrenceEnd = legacyRecurrenceEnd(wire.RecurrenceEnd, value.RecurrenceEnd)
	}
	value, err = domain.NormalizeLegacyTaskMetadata(value)
	if err != nil {
		return false, domain.LegacyTaskMetadata{}, ports.ErrInvalid
	}
	return true, value, nil
}

func legacyMetadataPatchPresent(wire writeWire) bool {
	return presentCompatibilityField(wire, "is_recurring") ||
		presentCompatibilityField(wire, "recurrence_rule") ||
		presentCompatibilityField(wire, "recurrence_end")
}

func legacyRecurrenceEnd(raw string, current *time.Time) *time.Time {
	if raw == "" {
		return nil
	}
	value, err := optionalTaskTime(raw)
	if err != nil {
		return current
	}
	return value
}
