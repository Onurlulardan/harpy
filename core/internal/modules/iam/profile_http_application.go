package iam

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

var (
	ErrProfileHTTPInvalid                = errors.New("IAM profile request is invalid")
	ErrProfileHTTPFieldForbidden         = errors.New("IAM profile field is forbidden")
	ErrProfileHTTPForbidden              = errors.New("IAM profile operation is forbidden")
	ErrProfileHTTPNotFound               = errors.New("IAM profile resource was not found")
	ErrProfileHTTPConflict               = errors.New("IAM profile operation conflicts")
	ErrProfileHTTPRecoveryRequired       = errors.New("IAM profile recovery is required")
	ErrProfileHTTPRecoveryBudgetExceeded = errors.New("IAM profile recovery budget is exceeded")
	ErrProfileHTTPUnavailable            = errors.New("IAM profile operation is unavailable")
)

type ProfileSelfUpdate struct {
	IdempotencyKey string
	Values         map[string]json.RawMessage
}

type ProfileValidationInput struct {
	MaximumBytes     int
	MaximumItems     int
	IntegerMinimum   *string
	IntegerMaximum   *string
	DecimalPrecision int
	DecimalScale     int
	EnumValues       []string
}

type ProfileFieldInput struct {
	ID             uuid.UUID
	Key            string
	Type           profileapp.LogicalType
	Cardinality    profileapp.Cardinality
	Required       bool
	Classification profileapp.Classification
	Normalization  profileapp.NormalizationProfile
	State          profileapp.FieldState
	Validation     ProfileValidationInput
}

type ProfileMigrationMappingInput struct {
	SourceFieldID uuid.UUID
	TargetFieldID uuid.UUID
	ConverterKind string
}

type ProfileSchemaPublish struct {
	IdempotencyKey         string
	SchemaID               uuid.UUID
	Resource               profileapp.ResourceKind
	Version                int64
	ExpectedActiveSchemaID uuid.UUID
	Fields                 []ProfileFieldInput
	MigrationMappings      []ProfileMigrationMappingInput
}

type ProfileSchemaRetire struct {
	IdempotencyKey         string
	SchemaID               uuid.UUID
	Resource               profileapp.ResourceKind
	ExpectedActiveSchemaID uuid.UUID
	FieldID                uuid.UUID
	ExpectedFieldRevision  int64
}

type ProfileSelfView struct {
	Document json.RawMessage
	Revision int64
}

type ProfileSchemaMutationView struct {
	ID                 uuid.UUID  `json:"id"`
	Revision           int64      `json:"revision"`
	AuthorizationClock int64      `json:"authorization_clock"`
	Replayed           bool       `json:"replayed"`
	BackfillPlanID     *uuid.UUID `json:"backfill_plan_id,omitempty"`
	PublicationState   string     `json:"publication_state,omitempty"`
}

type ProfileFieldView struct {
	ID               uuid.UUID                       `json:"id"`
	Key              string                          `json:"key"`
	Type             profileapp.LogicalType          `json:"type"`
	Cardinality      profileapp.Cardinality          `json:"cardinality"`
	Required         bool                            `json:"required"`
	Classification   profileapp.Classification       `json:"classification"`
	Normalization    profileapp.NormalizationProfile `json:"normalization"`
	State            profileapp.FieldState           `json:"state"`
	Validation       json.RawMessage                 `json:"validation"`
	ValidationSHA256 string                          `json:"validation_sha256"`
	SemanticSHA256   string                          `json:"semantic_sha256"`
}

type ProfileSchemaView struct {
	ID               uuid.UUID               `json:"id"`
	Resource         profileapp.ResourceKind `json:"resource"`
	Version          int64                   `json:"version"`
	State            profileapp.SchemaState  `json:"state"`
	Fields           []ProfileFieldView      `json:"fields"`
	Definition       json.RawMessage         `json:"definition"`
	DefinitionSHA256 string                  `json:"definition_sha256"`
}

type ProfileHTTPApplication interface {
	GetSelf(context.Context, accesstoken.Claims) (ProfileSelfView, error)
	UpdateSelf(
		context.Context, accesstoken.Claims, ProfileSelfUpdate,
	) (ProfileSelfView, error)
	GetActiveSchema(
		context.Context, accesstoken.Claims, profileapp.ResourceKind,
	) (ProfileSchemaView, error)
	GetSchema(
		context.Context, accesstoken.Claims, profileapp.ResourceKind, uuid.UUID,
	) (ProfileSchemaView, error)
	PublishSchema(
		context.Context, accesstoken.Claims, ProfileSchemaPublish,
	) (ProfileSchemaMutationView, error)
	RetireField(
		context.Context, accesstoken.Claims, ProfileSchemaRetire,
	) (ProfileSchemaMutationView, error)
	RecoverBackfill(
		context.Context, accesstoken.Claims, ProfileBackfillRecoveryInput,
	) (ProfileBackfillRecoveryView, error)
}
