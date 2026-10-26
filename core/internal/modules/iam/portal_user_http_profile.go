package iam

import (
	"context"
	"encoding/json"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *portalUserHTTPApplication) portalProfile(
	ctx context.Context, scope platformtx.TenantScope, firstName, lastName string,
) (identityfacade.ProfileDocument, error) {
	aliases, err := application.identities.profiles.FieldIDs(
		ctx, scope, identityfacade.ProfileUser, identityUserCoreProfileKeys,
	)
	if err != nil || len(aliases) != 2 {
		return identityfacade.ProfileDocument{}, ErrPortalUserHTTPUnavailable
	}
	values := make(map[string]json.RawMessage, 2)
	for key, value := range map[string]string{"first_name": firstName, "last_name": lastName} {
		id := uuid.UUID(aliases[key])
		encoded, marshalErr := json.Marshal(value)
		if id == uuid.Nil || marshalErr != nil {
			return identityfacade.ProfileDocument{}, ErrPortalUserHTTPInvalid
		}
		values[id.String()] = encoded
	}
	document, err := identityMarshalProfile(values)
	if err != nil {
		return identityfacade.ProfileDocument{}, ErrPortalUserHTTPInvalid
	}
	profile, err := application.identities.profiles.PrepareNew(
		ctx, scope, identityfacade.ProfileUser, document,
	)
	if err != nil || !portalProfileValid(profile) {
		return identityfacade.ProfileDocument{}, ErrPortalUserHTTPInvalid
	}
	return profile, nil
}
