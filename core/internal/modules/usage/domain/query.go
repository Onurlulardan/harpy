package domain

import (
	"regexp"
	"time"
)

const MaximumWindow = 31 * 24 * time.Hour

var (
	pseudonymPattern     = regexp.MustCompile(`^[0-9a-f]{64}$`)
	canonicalUUIDPattern = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
	)
)

type Scope string

const (
	TenantScope   Scope = "tenant"
	PlatformScope Scope = "platform"
)

type TimeRange struct{ Start, End time.Time }

func NewTimeRange(start, end time.Time) (TimeRange, error) {
	start, end = start.UTC(), end.UTC()
	if !storedTime(start) || !storedTime(end) || !start.Before(end) ||
		end.Sub(start) > MaximumWindow {
		return TimeRange{}, ErrInvalid
	}
	return TimeRange{Start: start, End: end}, nil
}

type RankedQuery struct {
	Range TimeRange
	Limit int
}

func NewRankedQuery(window TimeRange, limit, maximum int) (RankedQuery, error) {
	if window.Start.IsZero() || limit < 1 || limit > maximum {
		return RankedQuery{}, ErrInvalid
	}
	return RankedQuery{Range: window, Limit: limit}, nil
}

type TimelineQuery struct {
	Subject string
	Page    int
	Limit   int
}

func NewTimelineQuery(scope Scope, subject string, page, limit int) (TimelineQuery, error) {
	validSubject := pseudonymPattern.MatchString(subject)
	if scope == TenantScope {
		validSubject = canonicalUUIDPattern.MatchString(subject)
	}
	if !validSubject || page < 1 || page > 1000 || limit < 1 || limit > 100 {
		return TimelineQuery{}, ErrInvalid
	}
	return TimelineQuery{Subject: subject, Page: page, Limit: limit}, nil
}

type Granularity string

const (
	GranularityHour Granularity = "hour"
	GranularityDay  Granularity = "day"
	GranularityWeek Granularity = "week"
)

func ValidGranularity(value Granularity) bool {
	return value == GranularityHour || value == GranularityDay || value == GranularityWeek
}
