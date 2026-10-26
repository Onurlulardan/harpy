package actorbridgeissue

import (
	"context"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	actorbridgeissueapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/actorbridgeissue"
)

func (store *store) Resolve(
	ctx context.Context,
	scope actorbridgeissueapp.TenantScope,
	domain actorassertionv1.Domain,
) (actorbridgeissueapp.Evidence, error) {
	if ctx == nil || ctx.Err() != nil || domain.Validate() != nil {
		return actorbridgeissueapp.Evidence{}, errUnavailable
	}
	bound, err := store.bind(scope)
	if err != nil {
		return actorbridgeissueapp.Evidence{}, errUnavailable
	}
	var groupGeneration, iamGeneration, targetGeneration int64
	err = bound.database.QueryRow(ctx, resolveStatement,
		scope.TenantID(), string(domain)).Scan(
		&groupGeneration, &iamGeneration, &targetGeneration,
	)
	if err != nil {
		return actorbridgeissueapp.Evidence{}, err
	}
	return actorbridgeissueapp.NewEvidence(
		groupGeneration, iamGeneration, targetGeneration,
	)
}
