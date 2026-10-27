package lifecycle

import (
	"encoding/json"
	"sort"
	"time"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/google/uuid"
)

func encodeDemoRecordIDs(values []uuid.UUID) ([]byte, error) {
	canonical, err := canonicalDemoRecordIDs(values)
	if err != nil {
		return nil, err
	}
	wire := make([]string, len(canonical))
	for index, value := range canonical {
		wire[index] = value.String()
	}
	return json.Marshal(wire)
}

func decodeDemoRecordIDs(raw []byte) ([]uuid.UUID, error) {
	var wire []string
	if len(raw) == 0 || json.Unmarshal(raw, &wire) != nil {
		return nil, catalogapp.ErrUnavailable
	}
	values := make([]uuid.UUID, len(wire))
	for index, item := range wire {
		value, err := uuid.Parse(item)
		if err != nil || value == uuid.Nil || value.String() != item {
			return nil, catalogapp.ErrUnavailable
		}
		values[index] = value
	}
	canonical, err := canonicalDemoRecordIDs(values)
	if err != nil || !sameDemoUUIDs(canonical, values) {
		return nil, catalogapp.ErrUnavailable
	}
	return values, nil
}

func canonicalDemoRecordIDs(values []uuid.UUID) ([]uuid.UUID, error) {
	if len(values) > 10000 {
		return nil, catalogapp.ErrInvalid
	}
	result := append([]uuid.UUID(nil), values...)
	for _, value := range result {
		if value == uuid.Nil {
			return nil, catalogapp.ErrInvalid
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})
	for index := 1; index < len(result); index++ {
		if result[index-1] == result[index] {
			return nil, catalogapp.ErrInvalid
		}
	}
	return result, nil
}

func validDemoOperationPlan(value catalogapp.DemoOperationPlan) bool {
	if value.OperationID.Version() != 7 || value.InstallationID.Version() != 7 ||
		(value.Action != catalogapp.DemoActionInstall && value.Action != catalogapp.DemoActionRemove) ||
		len(value.Components) == 0 || !canonicalDemoTime(value.CreatedAt) {
		return false
	}
	if value.Action == catalogapp.DemoActionInstall && value.ExpectedPlanSHA256 != "" ||
		value.Action == catalogapp.DemoActionRemove && !digestPattern.MatchString(value.ExpectedPlanSHA256) {
		return false
	}
	total := 0
	for index, component := range value.Components {
		if component.ComponentID.Version() != 7 || component.SourceRunID.Version() != 7 {
			return false
		}
		ids, err := canonicalDemoRecordIDs(component.RecordIDs)
		if err != nil || !sameDemoUUIDs(ids, component.RecordIDs) ||
			value.Action == catalogapp.DemoActionRemove && len(ids) == 0 ||
			index > 0 && value.Components[index-1].ComponentID.String() >= component.ComponentID.String() {
			return false
		}
		total += len(ids)
	}
	return total <= 10000
}

func sameDemoUUIDs(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func canonicalDemoTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}
