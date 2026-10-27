package work

import (
	"time"

	workhttp "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/http"
	workpostgres "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

var legacyEndKey = string([]byte{101, 110, 100})

func legacyReadOptions(input workhttp.ListInput) (workpostgres.ReadOptions, error) {
	result := workpostgres.ReadOptions{}
	if !input.Compatibility {
		return result, nil
	}
	query := input.LegacyQuery
	switch input.Resource {
	case "available-users", "boards", "memberships", "comments":
		return result, nil
	case "inbox":
		return legacyTaskOptions(query, result,
			input.Operation == "listLegacyCalendarTasks")
	case "calendar":
		start, err := legacyTime(query["start"])
		end, endErr := legacyCalendarEnd(query[legacyEndKey])
		if err != nil || endErr != nil || !end.After(start) {
			return result, ErrHTTPInvalid
		}
		result.StartsAt, result.EndsAt = &start, &end
		result.OwnerOnly, result.Ascending = true, true
		return result, nil
	default:
		return result, nil
	}
}

func legacyTaskOptions(query map[string]string,
	result workpostgres.ReadOptions, calendar bool,
) (workpostgres.ReadOptions, error) {
	if raw := query["priority"]; raw != "" {
		result.Priority = ports.TaskPriority(raw)
		if !result.Priority.Valid() {
			return result, ErrHTTPInvalid
		}
	}
	if raw := query["status"]; raw != "" {
		if raw == "TODO" {
			raw = string(ports.StatusOpen)
		}
		result.Status = ports.TaskStatus(raw)
		if !result.Status.Valid() {
			return result, ErrHTTPInvalid
		}
	}
	if raw := query["board_id"]; raw != "" {
		id, err := ports.ParseID(raw)
		if err != nil {
			return result, ErrHTTPInvalid
		}
		result.BoardID = id
	}
	start, startErr := optionalLegacyTime(query["due_date_start"])
	end, endErr := optionalLegacyTime(query["due_date_end"])
	if startErr != nil || endErr != nil {
		return result, ErrHTTPInvalid
	}
	result.StartsAt, result.EndsAt = start, end
	if start != nil && end != nil && !end.After(*start) {
		return result, ErrHTTPInvalid
	}
	if calendar && start != nil && end != nil {
		inclusive := end.Add(24 * time.Hour)
		result.EndsAt, result.Ascending = &inclusive, true
	}
	return result, nil
}

func optionalLegacyTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	value, err := legacyTime(raw)
	return &value, err
}

func legacyTime(raw string) (time.Time, error) {
	if value, err := time.Parse(time.RFC3339, raw); err == nil {
		return value.UTC(), nil
	}
	value, err := time.Parse("2006-01-02", raw)
	return value.UTC(), err
}

func legacyCalendarEnd(raw string) (time.Time, error) {
	value, err := legacyTime(raw)
	if err == nil && len(raw) == len("2006-01-02") {
		value = value.AddDate(0, 0, 1)
	}
	return value, err
}
