package iam

import publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"

func identitySystemResourceCreateDigest(
	resource publicresource.Resource,
	values publicresource.FieldValues,
	passwordSHA [32]byte,
) ([32]byte, error) {
	valuesSHA, err := values.SHA256()
	if err != nil {
		return [32]byte{}, ErrIdentityHTTPInvalid
	}
	return identityRequestSHA(struct {
		Resource    string
		ValuesSHA   [32]byte
		PasswordSHA [32]byte
	}{string(resource), valuesSHA, passwordSHA})
}
