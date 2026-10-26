package resolved

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

const mandatoryVisibilitySchema = "restquery_mandatory_visibility_v1"
const maximumMandatoryCanonicalBytes = 64 * 1024

type mandatoryVisibilityDocument struct {
	Schema    string                      `json:"schema"`
	Resources []mandatoryResourceDocument `json:"resources"`
}

type mandatoryResourceDocument struct {
	RequirementSHA256  string          `json:"requirement_sha256"`
	Owner              string          `json:"owner"`
	ResourceID         string          `json:"resource_id"`
	PublishedVersionID string          `json:"published_version_id"`
	CatalogSHA256      string          `json:"catalog_sha256"`
	ResourcePath       string          `json:"resource_path"`
	Mode               string          `json:"mode"`
	Filter             json.RawMessage `json:"filter,omitempty"`
}

func canonicalMandatoryVisibilities(
	visibilities []ResourceVisibility,
) ([]byte, error) {
	document := mandatoryVisibilityDocument{
		Schema: mandatoryVisibilitySchema,
		Resources: make(
			[]mandatoryResourceDocument, 0, len(visibilities),
		),
	}
	for _, visibility := range visibilities {
		resource, err := mandatoryResourceCanonical(visibility)
		if err != nil {
			return nil, err
		}
		document.Resources = append(document.Resources, resource)
	}
	canonical, err := attestation.MarshalCanonical(document)
	if err != nil || len(canonical) < 1 ||
		len(canonical) > maximumMandatoryCanonicalBytes {
		return nil, fmt.Errorf("resolved REST Query mandatory visibility is invalid")
	}
	return canonical, nil
}

func mandatoryResourceCanonical(
	visibility ResourceVisibility,
) (mandatoryResourceDocument, error) {
	if visibility.Validate() != nil {
		return mandatoryResourceDocument{}, fmt.Errorf("invalid mandatory visibility")
	}
	owner, ownerErr := visibility.catalog.Owner()
	resourceID, resourceErr := visibility.catalog.ResourceID()
	versionID, versionErr := visibility.catalog.PublishedVersionID()
	hash, hashErr := visibility.catalog.CatalogHash()
	hashValue, hexErr := hash.Hex()
	path, pathErr := visibility.path.Value()
	if ownerErr != nil || resourceErr != nil || versionErr != nil ||
		hashErr != nil || hexErr != nil || pathErr != nil {
		return mandatoryResourceDocument{}, fmt.Errorf("invalid mandatory visibility")
	}
	document := mandatoryResourceDocument{
		RequirementSHA256: hex.EncodeToString(visibility.requirement[:]),
		Owner:             owner, ResourceID: resourceID.String(),
		PublishedVersionID: versionID.String(), CatalogSHA256: hashValue,
		ResourcePath: path, Mode: visibility.mode.String(),
	}
	if visibility.filter != nil {
		filter, err := canonicalMandatoryFilter(*visibility.filter)
		if err != nil {
			return mandatoryResourceDocument{}, err
		}
		document.Filter = filter
	}
	return document, nil
}

func sortCanonicalDocuments(documents []json.RawMessage) {
	slices.SortFunc(documents, func(left, right json.RawMessage) int {
		return bytes.Compare(left, right)
	})
}

func setValuedOperator(operator api.Operator) bool {
	return operator == api.OperatorIn || operator == api.OperatorNotIn ||
		operator == api.OperatorContainsAny || operator == api.OperatorContainsAll
}
