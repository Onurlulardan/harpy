package systemresource

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
)

func addProfileValues(
	values map[publicresource.ID]json.RawMessage,
	profile identityfacade.ProfileDocument,
	expected identityfacade.ProfileResource,
	visible map[publicresource.ID]struct{},
) error {
	if profile.Validate(expected) != nil {
		return publicresource.ErrContract
	}
	document := profile.Values()
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, identityfacade.MaximumProfileBytes,
	)
	if err != nil || !bytes.Equal(canonical, document) ||
		sha256.Sum256(canonical) != profile.SHA256() {
		return publicresource.ErrContract
	}
	var members map[string]json.RawMessage
	if json.Unmarshal(canonical, &members) != nil || members == nil {
		return publicresource.ErrContract
	}
	for key, value := range members {
		fieldID, parseErr := publicresource.ParseID(key)
		if parseErr != nil {
			return publicresource.ErrContract
		}
		if _, allowed := visible[fieldID]; !allowed {
			continue
		}
		if _, duplicate := values[fieldID]; duplicate {
			return publicresource.ErrContract
		}
		values[fieldID] = append(json.RawMessage(nil), value...)
	}
	return nil
}
