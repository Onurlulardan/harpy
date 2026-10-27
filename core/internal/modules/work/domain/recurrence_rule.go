package domain

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

const RecurrenceContract = "work_rrule_v1"

var ErrRecurrenceInvalid = errors.New("work recurrence rule is invalid")

type Frequency string

const (
	FrequencyDaily   Frequency = "DAILY"
	FrequencyWeekly  Frequency = "WEEKLY"
	FrequencyMonthly Frequency = "MONTHLY"
	FrequencyYearly  Frequency = "YEARLY"
)

type RecurrenceRule struct {
	Frequency  Frequency
	Interval   int
	Count      int
	Until      *time.Time
	ByDay      []time.Weekday
	ByMonthDay []int
	ByMonth    []time.Month
	WeekStart  time.Weekday
}

type RecurrenceDefinition struct {
	Contract, TimeZone, TZDBVersion string
	Start                           time.Time
	AllDay                          bool
	Rule                            RecurrenceRule
	Exceptions                      map[string]struct{}
}

type RecurrenceInput struct {
	DTStart, TimeZone, TZDBVersion string
	RRULE                          string
	EXDATE                         []string
	Location                       *time.Location
}

func ParseRecurrence(input RecurrenceInput) (RecurrenceDefinition, error) {
	if input.Location == nil || input.TimeZone == "" || input.TZDBVersion == "" ||
		input.Location.String() != input.TimeZone || len(input.EXDATE) > 500 {
		return RecurrenceDefinition{}, ErrRecurrenceInvalid
	}
	start, allDay, err := parseLocalValue(input.DTStart, input.Location)
	if err != nil {
		return RecurrenceDefinition{}, err
	}
	rule, err := parseRule(input.RRULE, input.Location)
	if err != nil {
		return RecurrenceDefinition{}, err
	}
	exceptions := make(map[string]struct{}, len(input.EXDATE))
	for _, value := range input.EXDATE {
		instant, exceptionAllDay, parseErr := parseLocalValue(value, input.Location)
		if parseErr != nil || exceptionAllDay != allDay {
			return RecurrenceDefinition{}, ErrRecurrenceInvalid
		}
		exceptions[occurrenceKey(instant, allDay)] = struct{}{}
	}
	return RecurrenceDefinition{Contract: RecurrenceContract, TimeZone: input.TimeZone,
		TZDBVersion: input.TZDBVersion, Start: start, AllDay: allDay,
		Rule: rule, Exceptions: exceptions}, nil
}

func parseRule(value string, location *time.Location) (RecurrenceRule, error) {
	rule := RecurrenceRule{Interval: 1, WeekStart: time.Monday}
	seen := map[string]bool{}
	for _, token := range strings.Split(value, ";") {
		parts := strings.SplitN(token, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" || seen[parts[0]] {
			return RecurrenceRule{}, ErrRecurrenceInvalid
		}
		seen[parts[0]] = true
		if err := applyRuleToken(&rule, parts[0], parts[1], location); err != nil {
			return RecurrenceRule{}, err
		}
	}
	if !rule.Frequency.valid() || rule.Interval < 1 || rule.Interval > 365 ||
		rule.Count < 0 || rule.Count > 10_000 || rule.Count > 0 && rule.Until != nil {
		return RecurrenceRule{}, ErrRecurrenceInvalid
	}
	return rule, nil
}

func applyRuleToken(rule *RecurrenceRule, key, value string, location *time.Location) error {
	switch key {
	case "FREQ":
		rule.Frequency = Frequency(value)
	case "INTERVAL":
		rule.Interval, _ = strconv.Atoi(value)
	case "COUNT":
		rule.Count, _ = strconv.Atoi(value)
	case "UNTIL":
		parsed, err := parseUntil(value, location)
		if err != nil {
			return err
		}
		rule.Until = &parsed
	case "BYDAY":
		rule.ByDay = parseWeekdays(value)
	case "BYMONTHDAY":
		rule.ByMonthDay = parseInts(value, -31, 31, true)
	case "BYMONTH":
		values := parseInts(value, 1, 12, false)
		for _, month := range values {
			rule.ByMonth = append(rule.ByMonth, time.Month(month))
		}
	case "WKST":
		values := parseWeekdays(value)
		if len(values) == 1 {
			rule.WeekStart = values[0]
		}
	default:
		return ErrRecurrenceInvalid
	}
	if key == "BYDAY" && len(rule.ByDay) == 0 || key == "BYMONTHDAY" && len(rule.ByMonthDay) == 0 ||
		key == "BYMONTH" && len(rule.ByMonth) == 0 || key == "WKST" && rule.WeekStart < 0 {
		return ErrRecurrenceInvalid
	}
	return nil
}

func (frequency Frequency) valid() bool {
	return frequency == FrequencyDaily || frequency == FrequencyWeekly ||
		frequency == FrequencyMonthly || frequency == FrequencyYearly
}
