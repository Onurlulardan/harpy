package api

import "time"

type Grant struct {
	Binding    Binding
	Revision   int64
	LeaseUntil time.Time
}

func (grant Grant) valid(now time.Time) bool {
	return grant.Binding.valid() && grant.Revision > 0 && !now.IsZero() &&
		now.Location() == time.UTC && grant.LeaseUntil.Location() == time.UTC &&
		grant.LeaseUntil.After(now)
}
