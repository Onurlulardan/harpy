package iam

import (
	"context"
	"sync"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

type identityPreparedSystemResourceCreate struct {
	mutex          sync.Mutex
	resource       publicresource.Resource
	values         publicresource.FieldValues
	idempotencyKey string
	credential     identityPreparedUserCredential
	valid          bool
}

func (port *identitySystemResourcePort) PrepareCreate(
	ctx context.Context,
	input publicresource.CreateInput,
) (publicresource.PreparedCreate, error) {
	defer input.Destroy()
	if port == nil || port.application == nil || ctx == nil || ctx.Err() != nil {
		return nil, publicresource.ErrContract
	}
	resource, resourceErr := input.Resource()
	values, valuesErr := input.Values()
	idempotencyKey, keyErr := input.IdempotencyKey()
	password, passwordErr := input.Password()
	if resourceErr != nil || valuesErr != nil || keyErr != nil || passwordErr != nil {
		clear(password)
		return nil, publicresource.ErrContract
	}
	prepared := &identityPreparedSystemResourceCreate{
		resource: resource, values: values,
		idempotencyKey: idempotencyKey, valid: true,
	}
	if resource == publicresource.Users {
		credential, err := port.application.prepareSystemUserCredential(
			ctx, values, password,
		)
		clear(password)
		if err != nil {
			prepared.Destroy()
			return nil, identitySystemResourceError(err)
		}
		prepared.credential = credential
		return prepared, nil
	}
	clear(password)
	return prepared, nil
}

func (prepared *identityPreparedSystemResourceCreate) Resource() (
	publicresource.Resource,
	error,
) {
	if prepared == nil {
		return "", publicresource.ErrContract
	}
	prepared.mutex.Lock()
	defer prepared.mutex.Unlock()
	if !prepared.valid || !prepared.resource.Valid() {
		return "", publicresource.ErrContract
	}
	return prepared.resource, nil
}

func (*identityPreparedSystemResourceCreate) Contract() string {
	return publicresource.PreparedCreateContract
}

func (prepared *identityPreparedSystemResourceCreate) Destroy() {
	if prepared == nil {
		return
	}
	prepared.mutex.Lock()
	defer prepared.mutex.Unlock()
	prepared.resource, prepared.values, prepared.idempotencyKey = "", publicresource.FieldValues{}, ""
	prepared.credential = identityPreparedUserCredential{}
	prepared.valid = false
}
