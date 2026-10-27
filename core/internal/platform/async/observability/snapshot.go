package observability

import (
	"context"
	"fmt"
)

type Snapshot struct {
	oldestUndeliveredOutboxAge int64
	undeliveredOutboxAlarm     int64
	oldestDueJobLag            int64
	parkedJobs                 int64
	parkedOutbox               int64
}

type Source interface {
	Snapshot(context.Context) (Snapshot, error)
}

func NewSnapshot(
	oldestUndeliveredOutboxAge int64,
	undeliveredOutboxAlarm int64,
	oldestDueJobLag int64,
	parkedJobs int64,
	parkedOutbox int64,
) (Snapshot, error) {
	if oldestUndeliveredOutboxAge < 0 || oldestDueJobLag < 0 ||
		parkedJobs < 0 || parkedOutbox < 0 ||
		(undeliveredOutboxAlarm != 0 && undeliveredOutboxAlarm != 1) {
		return Snapshot{}, fmt.Errorf("async observability snapshot is invalid")
	}
	return Snapshot{
		oldestUndeliveredOutboxAge: oldestUndeliveredOutboxAge,
		undeliveredOutboxAlarm:     undeliveredOutboxAlarm,
		oldestDueJobLag:            oldestDueJobLag,
		parkedJobs:                 parkedJobs,
		parkedOutbox:               parkedOutbox,
	}, nil
}

func (snapshot Snapshot) OldestUndeliveredOutboxAgeSeconds() int64 {
	return snapshot.oldestUndeliveredOutboxAge
}

func (snapshot Snapshot) UndeliveredOutboxAlarm() int64 {
	return snapshot.undeliveredOutboxAlarm
}

func (snapshot Snapshot) OldestDueJobLagSeconds() int64 {
	return snapshot.oldestDueJobLag
}

func (snapshot Snapshot) ParkedJobs() int64 {
	return snapshot.parkedJobs
}

func (snapshot Snapshot) ParkedOutbox() int64 {
	return snapshot.parkedOutbox
}
