package systemtransfer

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

type IdentityManifest struct {
	Resource publicresource.Resource
	Fields   []string
	Digest   [32]byte
}

func ManifestFor(resource publicresource.Resource) (IdentityManifest, error) {
	var fields []string
	switch resource {
	case publicresource.Users:
		fields = []string{publicresource.UserEmailFieldID}
	case publicresource.Roles:
		fields = []string{publicresource.RoleScopeKindFieldID,
			publicresource.RoleScopeIDFieldID, publicresource.RoleNameFieldID}
	case publicresource.Organizations:
		fields = []string{publicresource.OrganizationSlugFieldID}
	default:
		return IdentityManifest{}, fmt.Errorf("IAM transfer resource is invalid")
	}
	document, _ := json.Marshal(struct {
		Contract string                  `json:"contract"`
		Resource publicresource.Resource `json:"resource"`
		Fields   []string                `json:"fields"`
	}{Contract: "iam-system-resource-identity.v1", Resource: resource, Fields: fields})
	return IdentityManifest{Resource: resource, Fields: fields,
		Digest: sha256.Sum256(document)}, nil
}
