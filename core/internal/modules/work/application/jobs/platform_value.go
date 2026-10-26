package jobs

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

func NewPlatformInput(jobID, tenantID, targetID uuid.UUID, kind string,
	payloadSHA256 [32]byte, leaseGeneration int64, now time.Time,
) (Claim, error) {
	claim := Claim{JobID: domain.ID(jobID), TenantID: domain.ID(tenantID),
		TargetID: domain.ID(targetID), Kind: kind, PayloadSHA256: payloadSHA256,
		LeaseGeneration: leaseGeneration, Now: now}
	if !claim.Valid() {
		return Claim{}, ports.ErrInvalid
	}
	return claim, nil
}

func NewReminderInput(taskID, eventID uuid.UUID, taskVersion int64,
	dueAt time.Time, title string,
) (Reminder, error) {
	value := Reminder{TaskID: domain.ID(taskID), EventID: domain.ID(eventID),
		TaskVersion: taskVersion, DueAt: dueAt, Title: title}
	if !value.TaskID.Valid() || !value.EventID.UUIDv7() || value.TaskVersion < 1 ||
		value.DueAt.IsZero() || value.Title == "" {
		return Reminder{}, ports.ErrInvalid
	}
	return value, nil
}
