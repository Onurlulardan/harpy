package authorization

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

var catalogKeyPattern = regexp.MustCompile(
	`^[a-z][a-z0-9]*([-_][a-z0-9]+)*(\.[a-z][a-z0-9]*([-_][a-z0-9]+)*)*$`,
)

type Request struct {
	resourceKey          string
	actionKey            string
	tokenVersion         int64
	authorizationVersion int64
	actorID              uuid.UUID
	explicitActor        bool
	valid                bool
}

func newCurrentActorEvidenceRequest(
	resourceKey string,
	actionKey string,
	actorID uuid.UUID,
) (Request, error) {
	request, err := NewRequest(resourceKey, actionKey, 1, 1)
	if err != nil || actorID == uuid.Nil || actorID.Variant() != uuid.RFC4122 {
		return Request{}, fmt.Errorf("IAM authorization request is invalid")
	}
	request.actorID, request.explicitActor = actorID, true
	return request, nil
}

func NewRequest(
	resourceKey string,
	actionKey string,
	tokenVersion int64,
	authorizationVersion int64,
) (Request, error) {
	request := Request{
		resourceKey: resourceKey, actionKey: actionKey,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		valid: true,
	}
	if err := request.Validate(); err != nil {
		return Request{}, err
	}
	return request, nil
}

func (request Request) Validate() error {
	if !request.valid || !validCatalogKey(request.resourceKey) ||
		!validCatalogKey(request.actionKey) || request.tokenVersion <= 0 ||
		request.authorizationVersion <= 0 {
		return fmt.Errorf("IAM authorization request is invalid")
	}
	return nil
}

func (request Request) ResourceKey() string         { return request.resourceKey }
func (request Request) ActionKey() string           { return request.actionKey }
func (request Request) TokenVersion() int64         { return request.tokenVersion }
func (request Request) AuthorizationVersion() int64 { return request.authorizationVersion }
func (request Request) ActorID() (uuid.UUID, bool) {
	return request.actorID, request.explicitActor
}
func validCatalogKey(value string) bool {
	return len(value) <= 100 && catalogKeyPattern.MatchString(value)
}
