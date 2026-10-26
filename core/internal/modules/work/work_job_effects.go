package work

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobapp "github.com/Onurlulardan/harpy/core/internal/modules/work/application/jobs"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

type reminderEventPayload struct {
	DueAt       string `json:"due_at"`
	TaskID      string `json:"task_id"`
	TaskVersion int64  `json:"task_version"`
	Title       string `json:"title"`
}

func workReminderEvents(claim jobapp.Claim,
	reminders []jobapp.Reminder,
) ([]platformasync.DatabaseJobEventInput, error) {
	result := make([]platformasync.DatabaseJobEventInput, len(reminders))
	for index, reminder := range reminders {
		payload, digest, err := reminderPayload(reminder)
		if err != nil || !reminder.EventID.UUIDv7() {
			return nil, ports.ErrInvalid
		}
		key := "work:reminder:" + reminder.TaskID.String() + ":" +
			strconv.FormatInt(reminder.TaskVersion, 10)
		jobID := uuid.UUID(claim.JobID)
		result[index] = platformasync.DatabaseJobEventInput{
			EventID: uuid.UUID(reminder.EventID), AggregateID: uuid.UUID(reminder.TaskID),
			ActorID: jobID, CausationID: jobID,
			EventType:     asyncmanifestv1.WorkReminderDueEventType,
			AggregateType: "work.task", EventVersion: 1,
			AggregateVersion: reminder.TaskVersion, EventSequence: reminder.TaskVersion,
			OccurredAt: claim.Now, CorrelationID: jobID.String(),
			IdempotencyKey: key, BusinessEffectIdempotencyKey: key,
			PayloadSchemaSHA256: asyncmanifestv1.WorkReminderDueSchemaSHA256,
			PayloadSHA256:       digest, Payload: payload,
		}
	}
	return result, nil
}

func reminderPayload(reminder jobapp.Reminder) ([]byte, string, error) {
	dueAt := reminder.DueAt.UTC().Truncate(time.Microsecond)
	document, err := json.Marshal(reminderEventPayload{DueAt: dueAt.Format(time.RFC3339Nano),
		TaskID: reminder.TaskID.String(), TaskVersion: reminder.TaskVersion,
		Title: reminder.Title})
	if err != nil {
		return nil, "", err
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, eventsv1.MaximumCanonicalPayloadBytes,
	)
	if err != nil {
		return nil, "", err
	}
	digest := sha256.Sum256(canonical)
	return canonical, hex.EncodeToString(digest[:]), nil
}
