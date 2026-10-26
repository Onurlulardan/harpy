package work

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/tasks"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type legacyChecklistWire struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

func taskKind(wire writeWire, compatibility bool) (ports.TaskKind, error) {
	raw := wire.Kind
	if compatibility {
		raw = wire.TaskType
	}
	if raw == "" {
		return ports.TaskCustom, nil
	}
	kind := ports.TaskKind(raw)
	if !kind.Valid() {
		return "", ports.ErrInvalid
	}
	return kind, nil
}

func taskPriority(raw string, fallback ports.TaskPriority) (ports.TaskPriority, error) {
	if raw == "" {
		return fallback, nil
	}
	priority := ports.TaskPriority(raw)
	if !priority.Valid() {
		return "", ports.ErrInvalid
	}
	return priority, nil
}

func taskCreateTimes(wire writeWire, compatibility bool) (*time.Time, *time.Time, error) {
	starts, due := wire.StartsAt, wire.DueAt
	if compatibility {
		var startErr, dueErr error
		starts, startErr = optionalTaskTime(wire.StartDate)
		due, dueErr = optionalTaskTime(wire.DueDate)
		if startErr != nil || dueErr != nil {
			return nil, nil, ports.ErrInvalid
		}
	}
	return starts, due, nil
}

func optionalTaskTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	value = value.UTC()
	return &value, nil
}

func taskRelations(wire writeWire, compatibility, create bool) (tasks.RelationInput, error) {
	result := tasks.RelationInput{}
	if compatibility {
		result.ReplaceAssignees = create || presentCompatibilityField(wire, "assignee_id")
		result.ReplaceLabels = create || wire.Present["labels"]
		result.ReplaceChecklist = create || wire.Present["checklist"]
		principal, err := legacyAssignee(wire)
		if err != nil {
			return result, err
		}
		if principal != nil {
			result.Assignees = []ports.Principal{*principal}
		}
	} else {
		result.ReplaceAssignees = create || wire.Present["principals"]
		for _, value := range wire.Principals {
			id, err := parseID(value.ID)
			principal := ports.Principal{Kind: ports.PrincipalKind(value.Kind), ID: id}
			if err != nil || !principal.Valid() {
				return result, ports.ErrInvalid
			}
			result.Assignees = append(result.Assignees, principal)
		}
	}
	var err error
	if result.ReplaceLabels {
		result.Labels, err = decodeLabels(wire.Labels)
	}
	if err == nil && result.ReplaceChecklist {
		result.Checklist, err = decodeChecklist(wire.Checklist, create)
	}
	return result, err
}

func legacyAssignee(wire writeWire) (*ports.Principal, error) {
	if !presentCompatibilityField(wire, "assignee_id") {
		if presentCompatibilityField(wire, "assignee_type") {
			return nil, ports.ErrInvalid
		}
		return nil, nil
	}
	if wire.AssigneeID == "" {
		return nil, nil
	}
	id, err := parseID(wire.AssigneeID)
	kind := ports.PrincipalKind(wire.AssigneeType)
	if kind == "" {
		kind = ports.PrincipalUser
	}
	principal := ports.Principal{Kind: kind, ID: id}
	if err != nil || !principal.Valid() {
		return nil, ports.ErrInvalid
	}
	return &principal, nil
}

func decodeLabels(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	var values []string
	if json.Unmarshal(raw, &values) != nil {
		return nil, ports.ErrInvalid
	}
	return values, nil
}

func decodeChecklist(raw json.RawMessage, create bool) ([]tasks.ChecklistValue, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	var values []legacyChecklistWire
	if json.Unmarshal(raw, &values) != nil {
		return nil, ports.ErrInvalid
	}
	result := make([]tasks.ChecklistValue, 0, len(values))
	for _, value := range values {
		item := tasks.ChecklistValue{Text: value.Text, Completed: value.Checked}
		if !create && value.ID != "" && !strings.HasPrefix(value.ID, "cl-") {
			id, err := parseID(value.ID)
			if err != nil {
				return nil, ports.ErrInvalid
			}
			item.ID = &id
		}
		result = append(result, item)
	}
	return result, nil
}
