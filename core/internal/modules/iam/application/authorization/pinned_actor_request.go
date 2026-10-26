package authorization

import "fmt"

// PinnedActorRequest authorizes a durable actor-owned operation against the
// exact authorization version captured at submission. Token rotation does not
// invalidate the job; the current positive token version is returned for the
// nested command authority.
type PinnedActorRequest struct {
	resourceKey          string
	actionKey            string
	authorizationVersion int64
	valid                bool
}

func NewPinnedActorRequest(
	resourceKey string,
	actionKey string,
	authorizationVersion int64,
) (PinnedActorRequest, error) {
	request := PinnedActorRequest{resourceKey: resourceKey, actionKey: actionKey,
		authorizationVersion: authorizationVersion, valid: true}
	if err := request.Validate(); err != nil {
		return PinnedActorRequest{}, err
	}
	return request, nil
}

func (request PinnedActorRequest) Validate() error {
	if !request.valid || !validCatalogKey(request.resourceKey) ||
		!validCatalogKey(request.actionKey) || request.authorizationVersion < 1 {
		return fmt.Errorf("IAM pinned actor authorization request is invalid")
	}
	return nil
}

func (request PinnedActorRequest) ResourceKey() string { return request.resourceKey }
func (request PinnedActorRequest) ActionKey() string   { return request.actionKey }
func (request PinnedActorRequest) AuthorizationVersion() int64 {
	return request.authorizationVersion
}
