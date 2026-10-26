package work

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

var legacyTaskUpdateIntentFields = map[string]bool{
	"title": true, "description": true, "task_type": true, "priority": true,
	"status": true, "form_slug": true, "form_data_id": true,
	"assignee_id": true, "assignee_type": true, "due_date": true,
	"start_date": true, "labels": true, "checklist": true,
	"is_recurring": true, "recurrence_rule": true, "recurrence_end": true,
}

var legacyTaskUpdateNullablePointers = map[string]bool{
	"title": true, "description": true, "task_type": true, "priority": true,
	"status": true, "form_slug": true, "form_data_id": true,
	"assignee_id": true, "assignee_type": true, "due_date": true,
	"start_date": true, "is_recurring": true, "recurrence_rule": true,
	"recurrence_end": true,
}

// legacyTaskUpdateIntentHash fingerprints only fields accepted by the V1
// UpdateTaskRequest. It is deliberately independent from the current task row,
// row version, permission revisions, and fields silently ignored by V1.
func legacyTaskUpdateIntentHash(taskID ports.ID, body json.RawMessage) ([32]byte, error) {
	if !taskID.Valid() {
		return [32]byte{}, ports.ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var request map[string]any
	if decoder.Decode(&request) != nil || request == nil {
		return [32]byte{}, ports.ErrInvalid
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return [32]byte{}, ports.ErrInvalid
	}
	intent := make(map[string]any, len(request))
	for field, value := range request {
		if !legacyTaskUpdateIntentFields[field] || value == nil && legacyTaskUpdateNullablePointers[field] {
			continue
		}
		intent[field] = value
	}
	canonical, err := json.Marshal(intent)
	if err != nil {
		return [32]byte{}, ports.ErrInvalid
	}
	return ports.Fingerprint("work.legacy-task.update.intent.v1",
		taskID.String(), string(canonical)), nil
}
