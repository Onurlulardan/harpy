package domain

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

var weekdayCodes = map[string]time.Weekday{
	"SU": time.Sunday, "MO": time.Monday, "TU": time.Tuesday,
	"WE": time.Wednesday, "TH": time.Thursday, "FR": time.Friday,
	"SA": time.Saturday,
}

func parseWeekdays(value string) []time.Weekday {
	seen := map[time.Weekday]bool{}
	result := make([]time.Weekday, 0, 7)
	for _, part := range strings.Split(value, ",") {
		weekday, exists := weekdayCodes[part]
		if !exists || seen[weekday] {
			return nil
		}
		seen[weekday] = true
		result = append(result, weekday)
	}
	sort.Slice(result, func(left, right int) bool { return result[left] < result[right] })
	return result
}

func parseInts(value string, minimum, maximum int, rejectZero bool) []int {
	seen := map[int]bool{}
	result := make([]int, 0, len(strings.Split(value, ",")))
	for _, part := range strings.Split(value, ",") {
		parsed, err := strconv.Atoi(part)
		if err != nil || parsed < minimum || parsed > maximum || rejectZero && parsed == 0 || seen[parsed] {
			return nil
		}
		seen[parsed] = true
		result = append(result, parsed)
	}
	sort.Ints(result)
	return result
}

func parseLocalValue(value string, location *time.Location) (time.Time, bool, error) {
	if len(value) == 8 {
		parsed, err := time.ParseInLocation("20060102", value, location)
		if err != nil {
			return time.Time{}, false, ErrRecurrenceInvalid
		}
		return parsed, true, nil
	}
	if len(value) != 15 || value[8] != 'T' {
		return time.Time{}, false, ErrRecurrenceInvalid
	}
	label, err := time.Parse("20060102T150405", value)
	if err != nil {
		return time.Time{}, false, ErrRecurrenceInvalid
	}
	resolved, ok := resolveLocal(location, label.Year(), label.Month(), label.Day(),
		label.Hour(), label.Minute(), label.Second())
	if !ok {
		return time.Time{}, false, ErrRecurrenceInvalid
	}
	return resolved, false, nil
}

func parseUntil(value string, location *time.Location) (time.Time, error) {
	if strings.HasSuffix(value, "Z") {
		parsed, err := time.Parse("20060102T150405Z", value)
		if err != nil {
			return time.Time{}, ErrRecurrenceInvalid
		}
		return parsed.UTC(), nil
	}
	parsed, _, err := parseLocalValue(value, location)
	return parsed, err
}

func resolveLocal(location *time.Location, year int, month time.Month, day, hour, minute, second int) (time.Time, bool) {
	naive := time.Date(year, month, day, hour, minute, second, 0, time.UTC)
	offsets := map[int]struct{}{}
	for delta := -48; delta <= 48; delta++ {
		_, offset := naive.Add(time.Duration(delta) * time.Hour).In(location).Zone()
		offsets[offset] = struct{}{}
	}
	var candidates []time.Time
	for offset := range offsets {
		candidate := naive.Add(-time.Duration(offset) * time.Second)
		local := candidate.In(location)
		if local.Year() == year && local.Month() == month && local.Day() == day &&
			local.Hour() == hour && local.Minute() == minute && local.Second() == second {
			candidates = append(candidates, candidate)
		}
	}
	if len(candidates) == 0 {
		return time.Time{}, false
	}
	sort.Slice(candidates, func(left, right int) bool { return candidates[left].Before(candidates[right]) })
	return candidates[0].In(location), true
}

func occurrenceKey(value time.Time, allDay bool) string {
	if allDay {
		return value.Format("2006-01-02")
	}
	return value.UTC().Format(time.RFC3339Nano)
}
