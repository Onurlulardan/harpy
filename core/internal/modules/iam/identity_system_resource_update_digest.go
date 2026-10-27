package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

func identitySystemResourceUpdateDigest(
	resource publicresource.Resource,
	id uuid.UUID,
	expectedRevision int64,
	values publicresource.FieldValues,
) ([32]byte, error) {
	valuesSHA, err := values.SHA256()
	if err != nil {
		return [32]byte{}, ErrIdentityHTTPInvalid
	}
	return identityRequestSHA(struct {
		Resource         string
		ID               uuid.UUID
		ExpectedRevision int64
		ValuesSHA        [32]byte
	}{string(resource), id, expectedRevision, valuesSHA})
}
