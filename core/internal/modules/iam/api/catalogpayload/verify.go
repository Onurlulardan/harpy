package catalogpayload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

const MaximumPayloadBytes = 4 << 20

var (
	uuidV7Pattern = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	)
	resourcePattern = regexp.MustCompile(
		`^(module|form|workflow)\.[a-z][a-z0-9-]{1,62}$`,
	)
	actionPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,31}$`)
)

type Resource struct {
	Slug        string `json:"slug"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Grant struct {
	ResourceSlug    string   `json:"resource_slug"`
	RoleName        string   `json:"role_name"`
	RoleDescription string   `json:"role_description"`
	Actions         []string `json:"actions"`
}

type Payload struct {
	ContractVersion string     `json:"contract_version"`
	ComponentID     string     `json:"component_id"`
	ModuleSlug      string     `json:"module_slug"`
	SourceContract  string     `json:"source_contract"`
	Resources       []Resource `json:"resources"`
	Grants          []Grant    `json:"grants"`
}

type Verified struct {
	payload  Payload
	document []byte
}

func Verify(document []byte, componentID, expectedSHA256 string) (Verified, error) {
	if err := catalogpackagev1.ValidateCanonicalPayload(
		document, expectedSHA256, MaximumPayloadBytes,
	); err != nil {
		return Verified{}, fmt.Errorf("IAM catalog payload is invalid")
	}
	var payload Payload
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&payload) != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		validate(payload, componentID) != nil {
		return Verified{}, fmt.Errorf("IAM catalog payload contract is invalid")
	}
	return Verified{payload: clone(payload), document: append([]byte(nil), document...)}, nil
}

func validate(payload Payload, componentID string) error {
	expectedSource, sourceErr := catalogpackagev1.FirstPartySourceContract(
		payload.ModuleSlug,
	)
	if payload.ContractVersion != catalogpackagev1.ContractIAM ||
		payload.ComponentID != componentID ||
		!uuidV7Pattern.MatchString(payload.ComponentID) ||
		sourceErr != nil || payload.SourceContract != expectedSource ||
		len(payload.Resources) < 2 || len(payload.Grants) == 0 {
		return fmt.Errorf("IAM catalog identity is invalid")
	}
	resources := map[string]bool{}
	counts := map[string]int{}
	for _, resource := range payload.Resources {
		prefix := strings.SplitN(resource.Slug, ".", 2)[0]
		if !resourcePattern.MatchString(resource.Slug) || resource.Kind != prefix ||
			resource.Name == "" || resource.Description == "" || resources[resource.Slug] {
			return fmt.Errorf("IAM resource is invalid")
		}
		resources[resource.Slug], counts[prefix] = true, counts[prefix]+1
	}
	if counts["module"] != 1 || counts["form"]+counts["workflow"] < 1 {
		return fmt.Errorf("IAM resource composition is invalid")
	}
	if !resources["module."+payload.ModuleSlug] {
		return fmt.Errorf("IAM module resource is missing")
	}
	return validateGrants(payload.Grants, resources)
}

func validateGrants(grants []Grant, resources map[string]bool) error {
	seen, covered := map[string]bool{}, map[string]bool{}
	for _, grant := range grants {
		key := grant.ResourceSlug + "\x00" + grant.RoleName
		if !resources[grant.ResourceSlug] || grant.RoleName == "" ||
			grant.RoleDescription == "" || seen[key] || len(grant.Actions) == 0 {
			return fmt.Errorf("IAM grant is invalid")
		}
		seen[key], covered[grant.ResourceSlug] = true, true
		actions := map[string]bool{}
		for _, action := range grant.Actions {
			if !actionPattern.MatchString(action) || actions[action] {
				return fmt.Errorf("IAM grant action is invalid")
			}
			actions[action] = true
		}
	}
	if len(covered) != len(resources) {
		return fmt.Errorf("IAM resource grant coverage is incomplete")
	}
	return nil
}

func clone(payload Payload) Payload {
	payload.Resources = append([]Resource(nil), payload.Resources...)
	payload.Grants = append([]Grant(nil), payload.Grants...)
	for index := range payload.Grants {
		payload.Grants[index].Actions = append([]string(nil), payload.Grants[index].Actions...)
	}
	return payload
}

func (verified Verified) Value() Payload {
	if len(verified.document) == 0 {
		return Payload{}
	}
	return clone(verified.payload)
}

func (verified Verified) CanonicalJSON() []byte {
	return append([]byte(nil), verified.document...)
}
