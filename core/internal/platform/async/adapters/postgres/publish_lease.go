package postgres

import "time"

type PublishLease struct {
	heartbeatAt time.Time
	until       time.Time
}

func newPublishLease(heartbeatAt, until time.Time) (PublishLease, error) {
	if heartbeatAt.IsZero() || !until.Equal(heartbeatAt.Add(publishLeaseDuration)) {
		return PublishLease{}, ErrPublishClaimInvalid
	}
	return PublishLease{heartbeatAt: heartbeatAt, until: until}, nil
}

func (lease PublishLease) HeartbeatAt() time.Time { return lease.heartbeatAt }
func (lease PublishLease) Until() time.Time       { return lease.until }
