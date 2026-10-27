package adoption

import (
	"regexp"
	"time"
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var routePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._:/{}-]{0,255}$`)
var consumerPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,127}$`)

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}

func (application *Application) now() (time.Time, error) {
	value := application.clock.Now()
	if !canonicalTime(value) {
		return time.Time{}, ErrUnavailable
	}
	return value, nil
}
