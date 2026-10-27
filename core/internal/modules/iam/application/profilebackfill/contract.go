package profilebackfill

import (
	"crypto/sha256"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalid     = errors.New("IAM profile backfill plan is invalid")
	ErrConflict    = errors.New("IAM profile backfill recovery conflicts")
	ErrUnavailable = errors.New("IAM profile backfill plan is unavailable")
	ErrEffects     = errors.New("IAM profile backfill enqueue failed")
)

const (
	ConverterTypedMap = "typed_map_v1"
	MappingCopy       = "copy_v1"
	MappingNormalize  = "copy_normalize_v1"
)

type Evidence struct {
	OperationID                  uuid.UUID
	RequestSHA256                [sha256.Size]byte
	ActorID                      uuid.UUID
	ManagePermissionID           uuid.UUID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
}

type Mapping struct {
	SourceFieldID uuid.UUID
	TargetFieldID uuid.UUID
	ConverterKind string
}

type Page struct {
	JobID                uuid.UUID
	ScheduledRunAt       time.Time
	AfterResourceID      uuid.UUID
	ThroughResourceID    uuid.UUID
	ExpectedCount        int
	SourceSHA256         [sha256.Size]byte
	ExpectedTargetSHA256 [sha256.Size]byte
}

type TargetField struct {
	ID             uuid.UUID
	Required       bool
	Active         bool
	SemanticSHA256 [sha256.Size]byte
}

type TargetSchema struct {
	ID               uuid.UUID
	ResourceKind     string
	Version          int64
	DefinitionJSON   []byte
	DefinitionSHA256 [sha256.Size]byte
	Fields           []TargetField
}

type Plan struct {
	ID                   uuid.UUID
	SourceSchemaID       uuid.UUID
	Target               TargetSchema
	ConverterVersion     string
	SourceCount          int64
	PageSize             int
	InitialRollingSHA256 [sha256.Size]byte
	Mappings             []Mapping
	Pages                []Page
}

type CreateCommand struct {
	Plan     Plan
	Evidence Evidence
}

type Result struct {
	PlanID             uuid.UUID
	Revision           int64
	AuthorizationClock int64
	Replayed           bool
}

type FirstJob struct {
	PlanID           uuid.UUID
	JobID            uuid.UUID
	RunAt            time.Time
	PayloadReference string
	PayloadSHA256    string
	DedupeKey        string
}

type PreparedPlan struct {
	Command    CreateCommand
	Document   []byte
	PlanSHA256 [sha256.Size]byte
	FirstJob   FirstJob
}

func stableID(value uuid.UUID) bool {
	return value != uuid.Nil && value.Variant() == uuid.RFC4122
}
