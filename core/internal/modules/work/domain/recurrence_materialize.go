package domain

import (
	"errors"
	"time"
)

const (
	RecurrencePastWindow   = 24 * time.Hour
	RecurrenceFutureWindow = 180 * 24 * time.Hour
	MaxFuturePerRule       = 2_000
	MaxFuturePerTenant     = 50_000
)

var ErrRecurrenceLimit = errors.New("work recurrence limit is exceeded")

type Occurrence struct {
	Start       time.Time
	LocalLabel  string
	UTCOffset   int
	TZDBVersion string
}

func Materialize(definition RecurrenceDefinition, from, through time.Time, limit int) ([]Occurrence, error) {
	if definition.Contract != RecurrenceContract || !definition.Rule.Frequency.valid() ||
		definition.Start.IsZero() || definition.TZDBVersion == "" || !through.After(from) ||
		limit < 1 || limit > MaxFuturePerRule {
		return nil, ErrRecurrenceInvalid
	}
	location := definition.Start.Location()
	base := localMidnight(definition.Start)
	end := localMidnight(through.In(location)).AddDate(0, 0, 1)
	result := make([]Occurrence, 0)
	matched := 0
	for date := base; date.Before(end); date = date.AddDate(0, 0, 1) {
		if !ruleMatchesDate(definition, date) {
			continue
		}
		instant, ok := occurrenceAt(definition, date)
		if !ok {
			continue
		}
		if definition.Rule.Until != nil && instant.After(*definition.Rule.Until) {
			break
		}
		matched++
		if definition.Rule.Count > 0 && matched > definition.Rule.Count {
			break
		}
		if instant.Before(from) || !instant.Before(through) {
			continue
		}
		if _, excluded := definition.Exceptions[occurrenceKey(instant, definition.AllDay)]; excluded {
			continue
		}
		if len(result) == limit {
			return nil, ErrRecurrenceLimit
		}
		_, offset := instant.Zone()
		result = append(result, Occurrence{Start: instant.UTC(),
			LocalLabel: localLabel(instant, definition.AllDay), UTCOffset: offset,
			TZDBVersion: definition.TZDBVersion})
	}
	return result, nil
}

func ruleMatchesDate(definition RecurrenceDefinition, date time.Time) bool {
	rule := definition.Rule
	base := localMidnight(definition.Start)
	days := civilDays(base, date)
	months := (date.Year()-base.Year())*12 + int(date.Month()-base.Month())
	match := false
	switch rule.Frequency {
	case FrequencyDaily:
		match = days >= 0 && days%rule.Interval == 0
	case FrequencyWeekly:
		weekBase := base.AddDate(0, 0, -weekdayDistance(rule.WeekStart, base.Weekday()))
		weeks := civilDays(weekBase, date) / 7
		match = days >= 0 && weeks%rule.Interval == 0 &&
			containsWeekday(defaultWeekdays(rule.ByDay, base.Weekday()), date.Weekday())
	case FrequencyMonthly:
		match = months >= 0 && months%rule.Interval == 0 && matchesMonthDay(rule, base, date)
	case FrequencyYearly:
		match = date.Year() >= base.Year() && (date.Year()-base.Year())%rule.Interval == 0 &&
			matchesMonth(rule, base, date) && matchesMonthDay(rule, base, date)
	}
	return match && (len(rule.ByDay) == 0 || containsWeekday(rule.ByDay, date.Weekday())) &&
		(len(rule.ByMonth) == 0 || containsMonth(rule.ByMonth, date.Month()))
}

func occurrenceAt(definition RecurrenceDefinition, date time.Time) (time.Time, bool) {
	if definition.AllDay {
		return date, true
	}
	start := definition.Start.In(date.Location())
	return resolveLocal(date.Location(), date.Year(), date.Month(), date.Day(),
		start.Hour(), start.Minute(), start.Second())
}
