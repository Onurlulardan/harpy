package authorization

import (
	"fmt"

	"github.com/google/uuid"
)

type CurrentActorRequest struct {
	resourceKey string
	actionKey   string
	actorID     uuid.UUID
	valid       bool
}

func NewCurrentActorRequest(
	resourceKey string,
	actionKey string,
	actorID uuid.UUID,
) (CurrentActorRequest, error) {
	request := CurrentActorRequest{
		resourceKey: resourceKey, actionKey: actionKey,
		actorID: actorID, valid: true,
	}
	if err := request.Validate(); err != nil {
		return CurrentActorRequest{}, err
	}
	return request, nil
}

func (request CurrentActorRequest) Validate() error {
	if !request.valid || !validCatalogKey(request.resourceKey) ||
		!validCatalogKey(request.actionKey) || request.actorID == uuid.Nil ||
		request.actorID.Variant() != uuid.RFC4122 {
		return fmt.Errorf("IAM current actor request is invalid")
	}
	return nil
}

func (request CurrentActorRequest) ResourceKey() string { return request.resourceKey }
func (request CurrentActorRequest) ActionKey() string   { return request.actionKey }
func (request CurrentActorRequest) ActorID() uuid.UUID  { return request.actorID }
