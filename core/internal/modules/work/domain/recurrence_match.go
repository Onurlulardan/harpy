package domain

import "time"

func localMidnight(value time.Time) time.Time {
	local := value.In(value.Location())
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
}

func civilDays(left, right time.Time) int {
	leftUTC := time.Date(left.Year(), left.Month(), left.Day(), 0, 0, 0, 0, time.UTC)
	rightUTC := time.Date(right.Year(), right.Month(), right.Day(), 0, 0, 0, 0, time.UTC)
	return int(rightUTC.Sub(leftUTC) / (24 * time.Hour))
}

func weekdayDistance(start, value time.Weekday) int {
	result := int(value - start)
	if result < 0 {
		result += 7
	}
	return result
}

func defaultWeekdays(values []time.Weekday, fallback time.Weekday) []time.Weekday {
	if len(values) == 0 {
		return []time.Weekday{fallback}
	}
	return values
}

func containsWeekday(values []time.Weekday, wanted time.Weekday) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func containsMonth(values []time.Month, wanted time.Month) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func matchesMonth(rule RecurrenceRule, base, date time.Time) bool {
	if len(rule.ByMonth) == 0 {
		return date.Month() == base.Month()
	}
	return containsMonth(rule.ByMonth, date.Month())
}

func matchesMonthDay(rule RecurrenceRule, base, date time.Time) bool {
	if len(rule.ByMonthDay) == 0 {
		return date.Day() == base.Day()
	}
	daysInMonth := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location()).Day()
	for _, value := range rule.ByMonthDay {
		day := value
		if value < 0 {
			day = daysInMonth + value + 1
		}
		if date.Day() == day {
			return true
		}
	}
	return false
}

func localLabel(value time.Time, allDay bool) string {
	if allDay {
		return value.Format("2006-01-02")
	}
	return value.Format("2006-01-02T15:04:05")
}
