package work

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type principalWire struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type writeWire struct {
	Present             map[string]bool `json:"-"`
	Null                map[string]bool `json:"-"`
	ExpectedVersion     int64           `json:"expected_version"`
	ExpectedRuleVersion int64           `json:"expected_rule_revision"`
	TaskVersion         int64           `json:"task_version"`
	Name                string          `json:"name"`
	Description         string          `json:"description"`
	InitialColumn       string          `json:"initial_column"`
	BoardID             string          `json:"board_id"`
	ColumnID            string          `json:"column_id"`
	TaskID              string          `json:"task_id"`
	EventID             string          `json:"event_id"`
	MembershipID        string          `json:"membership_id"`
	PrincipalKind       string          `json:"principal_kind"`
	PrincipalID         string          `json:"principal_id"`
	UserID              string          `json:"user_id"`
	Role                string          `json:"role"`
	Active              *bool           `json:"active"`
	Title               string          `json:"title"`
	Kind                string          `json:"task_kind"`
	DueAt               *time.Time      `json:"due_at"`
	DueDate             string          `json:"due_date"`
	StartDate           string          `json:"start_date"`
	EndDate             string          `json:"end_date"`
	AllDay              *bool           `json:"all_day"`
	TaskType            string          `json:"task_type"`
	Priority            string          `json:"priority"`
	AssigneeID          string          `json:"assignee_id"`
	AssigneeType        string          `json:"assignee_type"`
	Order               int             `json:"order"`
	FormSlug            string          `json:"form_slug"`
	FormDataID          string          `json:"form_data_id"`
	FormFilter          json.RawMessage `json:"form_filter"`
	Labels              json.RawMessage `json:"labels"`
	Checklist           json.RawMessage `json:"checklist"`
	IsRecurring         *bool           `json:"is_recurring"`
	RecurrenceRule      string          `json:"recurrence_rule"`
	RecurrenceEnd       string          `json:"recurrence_end"`
	BeforeID            string          `json:"before_id"`
	AfterID             string          `json:"after_id"`
	Principals          []principalWire `json:"principals"`
	Body                string          `json:"body"`
	Content             string          `json:"content"`
	Reason              string          `json:"reason"`
	Color               string          `json:"color"`
	Linked              *bool           `json:"linked"`
	Text                string          `json:"text"`
	Completed           *bool           `json:"completed"`
	Position            int64           `json:"position"`
	Status              string          `json:"status"`
	LinkedStatus        string          `json:"linked_status"`
	Audience            string          `json:"audience"`
	OrganizationID      string          `json:"organization_id"`
	TimeKind            string          `json:"time_kind"`
	Timezone            string          `json:"timezone"`
	StartsAt            *time.Time      `json:"starts_at"`
	EndsAt              *time.Time      `json:"ends_at"`
	StartsOn            string          `json:"starts_on"`
	EndsOn              string          `json:"ends_on"`
	DTSTART             string          `json:"dtstart"`
	TZDBVersion         string          `json:"tzdb_version"`
	RRULE               string          `json:"rrule"`
	EXDATE              []string        `json:"exdate"`
	ColumnIDs           []string        `json:"column_ids"`
}

func decodeWrite(body json.RawMessage) (writeWire, error) {
	var value writeWire
	var fields map[string]json.RawMessage
	if json.Unmarshal(body, &fields) != nil || fields == nil {
		return writeWire{}, ports.ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return writeWire{}, ports.ErrInvalid
	}
	value.Present = make(map[string]bool, len(fields))
	value.Null = make(map[string]bool, len(fields))
	for field, raw := range fields {
		value.Present[field] = true
		value.Null[field] = bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
	}
	return value, nil
}

func presentCompatibilityField(wire writeWire, field string) bool {
	return wire.Present[field] && !wire.Null[field]
}

func parseID(value string) (ports.ID, error) {
	if value == "" {
		return ports.ID{}, ports.ErrInvalid
	}
	id, err := ports.ParseID(value)
	if err != nil {
		return ports.ID{}, ports.ErrInvalid
	}
	return id, nil
}

func optionalID(value string) (*ports.ID, error) {
	if value == "" {
		return nil, nil
	}
	id, err := parseID(value)
	return &id, err
}

func mutation(input HTTPWriteInput, version int64) ports.Mutation {
	return ports.Mutation{IdempotencyKey: input.IdempotencyKey,
		ExpectedVersion: version}
}
