package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

type includePlanBinding struct {
	tenant        uuid.UUID
	actor         uuid.UUID
	authorization [sha256.Size]byte
	mandatory     [sha256.Size]byte
	identity      [sha256.Size]byte
	rootCatalog   api.CatalogReference
	validated     bool
}

func newIncludePlanBinding(
	plan resolved.AuthorizedPlan,
) (includePlanBinding, error) {
	if plan.Validate() != nil {
		return includePlanBinding{}, ErrContract
	}
	spec, specErr := plan.Spec()
	tenant, tenantErr := plan.TenantID()
	actor, actorErr := plan.ActorID()
	authorization, authorizationErr := plan.AuthorizationContextSHA256()
	mandatory, mandatoryErr := plan.MandatoryCanonical()
	identity, identityErr := spec.IdentityHash()
	reference, referenceErr := spec.Reference()
	if specErr != nil || tenantErr != nil || actorErr != nil ||
		authorizationErr != nil || mandatoryErr != nil || identityErr != nil ||
		referenceErr != nil || len(mandatory) == 0 {
		return includePlanBinding{}, ErrContract
	}
	binding := includePlanBinding{
		tenant: tenant, actor: actor, authorization: authorization,
		mandatory: sha256.Sum256(mandatory), identity: identity,
		rootCatalog: reference, validated: true,
	}
	if binding.validate() != nil {
		return includePlanBinding{}, ErrContract
	}
	return binding, nil
}

func (binding includePlanBinding) validate() error {
	if !binding.validated || binding.tenant == uuid.Nil || binding.actor == uuid.Nil ||
		binding.tenant.Variant() != uuid.RFC4122 ||
		binding.actor.Variant() != uuid.RFC4122 ||
		binding.authorization == [sha256.Size]byte{} ||
		binding.mandatory == [sha256.Size]byte{} ||
		binding.identity == [sha256.Size]byte{} ||
		binding.rootCatalog.Validate() != nil {
		return ErrContract
	}
	return nil
}
