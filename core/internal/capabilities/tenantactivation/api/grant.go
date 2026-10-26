package api

import (
	"regexp"
	"time"
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Grant struct {
	Binding    Binding
	Revision   int64
	LeaseUntil time.Time
}

func (grant Grant) valid(now time.Time) bool {
	return grant.Binding.valid() && grant.Revision > 0 && !grant.LeaseUntil.IsZero() &&
		grant.LeaseUntil.After(now)
}
