package systemresources

import (
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

// The HTTP edge consumes the canonical orchestration contract through this
// facade. IAM's internal system-resource package remains behind orchestration.
type Resource = publicresource.Resource
type ID = publicresource.ID
type ActorEvidence = publicresource.ActorEvidence
type FieldCatalog = publicresource.FieldCatalog
type FieldValues = publicresource.FieldValues
type RelationProjection = publicresource.RelationProjection
type RelationValue = publicresource.RelationValue
type RelationSummary = publicresource.RelationSummary
type CreateInput = publicresource.CreateInput
type PreparedCreate = publicresource.PreparedCreate
type GetInput = publicresource.GetInput
type QueryRequest = publicresource.QueryRequest
type QueryInput = publicresource.QueryInput
type UpdateInput = publicresource.UpdateInput
type DeleteInput = publicresource.DeleteInput
type StateInput = publicresource.StateInput
type ReadRecord = publicresource.ReadResult
type QueryRecords = publicresource.QueryResult
type LayoutAccessRequest = publicresource.LayoutAccessRequest
type LayoutAccessDecision = publicresource.LayoutAccessDecision

const (
	Users         = publicresource.Users
	Roles         = publicresource.Roles
	Organizations = publicresource.Organizations
)

var (
	ErrLayoutAccessAuthentication = publicresource.ErrLayoutAccessAuthentication
	ErrLayoutAccessForbidden      = publicresource.ErrLayoutAccessForbidden
	ErrLayoutAccessUnavailable    = publicresource.ErrLayoutAccessUnavailable
)

func ParseID(value string) (ID, error) {
	return publicresource.ParseID(value)
}

func NewActorEvidence(
	tenantID ID,
	actorID ID,
	tokenVersion int64,
	authorizationVersion int64,
) (ActorEvidence, error) {
	return publicresource.NewActorEvidence(
		tenantID, actorID, tokenVersion, authorizationVersion,
	)
}

func NewLayoutAccessDecision(
	admission Admission,
) (LayoutAccessDecision, error) {
	return publicresource.NewLayoutAccessDecision(admission)
}

func NewFieldValues(values map[ID]json.RawMessage) (FieldValues, error) {
	return publicresource.NewFieldValues(values)
}

func NewRelationSummary(
	id ID,
	displayName string,
	details map[string]string,
) (RelationSummary, error) {
	return publicresource.NewRelationSummary(id, displayName, details)
}

func NewToOneRelation(
	field string,
	summary *RelationSummary,
) (RelationValue, error) {
	return publicresource.NewToOneRelation(field, summary)
}

func NewToManyRelation(
	field string,
	summaries []RelationSummary,
) (RelationValue, error) {
	return publicresource.NewToManyRelation(field, summaries)
}

func NewRelationProjection(
	resource Resource,
	values []RelationValue,
) (RelationProjection, error) {
	return publicresource.NewRelationProjection(resource, values)
}

func NewCreateInput(
	resource Resource,
	values FieldValues,
	password []byte,
	idempotencyKey string,
) (CreateInput, error) {
	return publicresource.NewCreateInput(resource, values, password, idempotencyKey)
}

func NewGetInput(resource Resource, id ID) (GetInput, error) {
	return publicresource.NewGetInput(resource, id)
}

func NewQueryInput(resource Resource, request QueryRequest) (
	QueryInput,
	error,
) {
	return publicresource.NewQueryInput(resource, request)
}

func NewUpdateInput(
	resource Resource,
	id ID,
	expectedRevision int64,
	values FieldValues,
	idempotencyKey string,
) (UpdateInput, error) {
	return publicresource.NewUpdateInput(
		resource, id, expectedRevision, values, idempotencyKey,
	)
}

func NewDeleteInput(
	resource Resource,
	id ID,
	expectedRevision int64,
	idempotencyKey string,
) (DeleteInput, error) {
	return publicresource.NewDeleteInput(
		resource, id, expectedRevision, idempotencyKey,
	)
}
