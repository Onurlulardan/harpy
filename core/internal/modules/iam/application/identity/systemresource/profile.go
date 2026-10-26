package systemresource

import identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"

type ProfileResource = identitydomain.ProfileResource
type ProfileDocument = identitydomain.ProfileDocument

const (
	ProfileUser         = identitydomain.ProfileUser
	ProfileRole         = identitydomain.ProfileRole
	ProfileOrganization = identitydomain.ProfileOrganization
	MaximumProfileBytes = identitydomain.MaximumProfileBytes
)

func NewProfileDocument(
	resource ProfileResource,
	schemaID IdentityID,
	values []byte,
	valuesSHA256 [32]byte,
) (ProfileDocument, error) {
	return identitydomain.NewProfileDocument(resource, schemaID, values, valuesSHA256)
}
