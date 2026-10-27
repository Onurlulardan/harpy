package systemresource

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
)

func publicResource(
	published catalog.Published,
) (publicresource.Resource, error) {
	resource, err := published.Resource()
	if err != nil {
		return "", publicresource.ErrContract
	}
	switch resource.Kind() {
	case profile.ResourceUser:
		return publicresource.Users, nil
	case profile.ResourceRole:
		return publicresource.Roles, nil
	case profile.ResourceOrganization:
		return publicresource.Organizations, nil
	default:
		return "", publicresource.ErrContract
	}
}

func catalogReference(
	published catalog.Published,
) (publicresource.ID, publicresource.ID, [32]byte, error) {
	reference, err := published.Reference()
	if err != nil {
		return publicresource.ID{}, publicresource.ID{}, [32]byte{}, publicresource.ErrContract
	}
	owner, ownerErr := reference.Owner()
	resourceID, resourceErr := reference.ResourceID()
	versionID, versionErr := reference.PublishedVersionID()
	hash, hashErr := reference.CatalogHash()
	bytes, bytesErr := hash.Bytes()
	if ownerErr != nil || owner != catalog.Owner || resourceErr != nil ||
		versionErr != nil || hashErr != nil || bytesErr != nil || len(bytes) != 32 {
		return publicresource.ID{}, publicresource.ID{}, [32]byte{}, publicresource.ErrContract
	}
	var digest [32]byte
	copy(digest[:], bytes)
	return publicresource.ID(resourceID), publicresource.ID(versionID), digest, nil
}
