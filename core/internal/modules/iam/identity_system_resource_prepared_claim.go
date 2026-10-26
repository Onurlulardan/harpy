package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

type identitySystemResourceCreateClaim struct {
	resource       publicresource.Resource
	values         publicresource.FieldValues
	idempotencyKey string
	credential     identityPreparedUserCredential
}

func claimIdentitySystemResourceCreate(
	input publicresource.PreparedCreate,
) (identitySystemResourceCreateClaim, error) {
	prepared, ok := input.(*identityPreparedSystemResourceCreate)
	if !ok || prepared == nil || prepared.Contract() != publicresource.PreparedCreateContract {
		return identitySystemResourceCreateClaim{}, publicresource.ErrContract
	}
	prepared.mutex.Lock()
	defer prepared.mutex.Unlock()
	if !prepared.valid || !prepared.resource.Valid() ||
		prepared.values.Validate() != nil || prepared.idempotencyKey == "" {
		return identitySystemResourceCreateClaim{}, publicresource.ErrContract
	}
	claim := identitySystemResourceCreateClaim{
		resource: prepared.resource, values: prepared.values,
		idempotencyKey: prepared.idempotencyKey, credential: prepared.credential,
	}
	prepared.resource, prepared.values, prepared.idempotencyKey = "", publicresource.FieldValues{}, ""
	prepared.credential, prepared.valid = identityPreparedUserCredential{}, false
	return claim, nil
}
