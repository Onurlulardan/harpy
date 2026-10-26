package recordkernel

import (
	"crypto/sha256"
	"encoding/json"
	"slices"

	"github.com/google/uuid"
)

type Action string

const (
	ActionCreate Action = "create"
	ActionGet    Action = "get"
	ActionList   Action = "list"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

type AuthorityKind uint8

const (
	AuthorityActor AuthorityKind = iota + 1
	AuthorityWorkflow
)

type MutationAuthority struct{ kind AuthorityKind }

func NewActorAuthority() MutationAuthority {
	return MutationAuthority{kind: AuthorityActor}
}

func NewWorkflowAuthority() MutationAuthority {
	return MutationAuthority{kind: AuthorityWorkflow}
}

func (authority MutationAuthority) Kind() AuthorityKind { return authority.kind }

func (authority MutationAuthority) Valid() bool {
	return authority.kind == AuthorityActor || authority.kind == AuthorityWorkflow
}

type Requirement struct {
	action        Action
	publicationID uuid.UUID
	formID        uuid.UUID
	formVersionID uuid.UUID
	fieldIDs      []uuid.UUID
}

func NewRequirement(action Action, publicationID, formID, versionID uuid.UUID,
	fieldIDs []uuid.UUID) (Requirement, error) {
	fields := append([]uuid.UUID(nil), fieldIDs...)
	slices.SortFunc(fields, compareUUID)
	for index, fieldID := range fields {
		if !stableUUIDv7(fieldID) || index > 0 && fieldID == fields[index-1] {
			return Requirement{}, ErrInputInvalid
		}
	}
	if !actionValid(action) || !stableUUIDv7(publicationID) ||
		!stableUUIDv7(formID) || !stableUUIDv7(versionID) {
		return Requirement{}, ErrInputInvalid
	}
	return Requirement{action, publicationID, formID, versionID, fields}, nil
}

func (value Requirement) Action() Action           { return value.action }
func (value Requirement) PublicationID() uuid.UUID { return value.publicationID }
func (value Requirement) FormID() uuid.UUID        { return value.formID }
func (value Requirement) FormVersionID() uuid.UUID { return value.formVersionID }
func (value Requirement) Fields() []uuid.UUID {
	return append([]uuid.UUID(nil), value.fieldIDs...)
}

func (value Requirement) SHA256() ([32]byte, error) {
	fields := make([]string, len(value.fieldIDs))
	for index, fieldID := range value.fieldIDs {
		fields[index] = fieldID.String()
	}
	document := struct {
		Action        Action   `json:"action"`
		PublicationID string   `json:"publication_id"`
		FormID        string   `json:"form_id"`
		FormVersionID string   `json:"form_version_id"`
		FieldIDs      []string `json:"field_ids"`
	}{value.action, value.publicationID.String(), value.formID.String(),
		value.formVersionID.String(), fields}
	encoded, err := json.Marshal(document)
	if err != nil {
		return [32]byte{}, ErrInputInvalid
	}
	return sha256.Sum256(encoded), nil
}

func actionValid(action Action) bool {
	return action == ActionCreate || action == ActionGet || action == ActionList ||
		action == ActionUpdate || action == ActionDelete
}
