package analytics

import "time"

type compatibilityClock struct{}

func (compatibilityClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}
