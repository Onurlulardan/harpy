package lifecycle

import (
	"regexp"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)

func validPins(value Plan) bool {
	dependencies := make(map[string]bool, len(value.Dependencies))
	for _, pin := range value.Dependencies {
		key := pin.PackageID.String()
		if !pin.PackageID.Valid() || !pin.Version.Valid() || dependencies[key] ||
			key == value.PackageID.String() {
			return false
		}
		dependencies[key] = true
	}
	components := make(map[string]bool, len(value.Components))
	for _, pin := range value.Components {
		key := pin.ComponentID.String()
		if !pin.ComponentID.Valid() || components[key] || !validOwner(pin.Owner) ||
			pin.OwnerGeneration <= 0 || pin.Contract == "" || len(pin.Contract) > 64 ||
			!digestPattern.MatchString(pin.ContentSHA256) {
			return false
		}
		components[key] = true
	}
	menuItems := make(map[string]bool, len(value.Menu))
	for _, pin := range value.Menu {
		key := pin.ItemID.String()
		if !pin.ItemID.Valid() || menuItems[key] || pin.Label == "" || len(pin.Label) > 128 ||
			pin.Order < 0 || pin.Order > 100000 || !validMenuPin(pin) {
			return false
		}
		menuItems[key] = true
	}
	for _, pin := range value.Menu {
		if pin.ParentID != nil && !menuItems[pin.ParentID.String()] {
			return false
		}
	}
	return true
}

func validMenuPin(value MenuPin) bool {
	switch value.Kind {
	case packagecontract.MenuGroup:
		return value.TargetID == nil && value.URL == ""
	case packagecontract.MenuExternalLink:
		return value.TargetID == nil && len(value.URL) > len("https://") &&
			strings.HasPrefix(value.URL, "https://")
	case packagecontract.MenuForm, packagecontract.MenuWorkflow,
		packagecontract.MenuDashboard, packagecontract.MenuEndpoint:
		return value.TargetID != nil && value.TargetID.Valid() && value.URL == ""
	default:
		return false
	}
}

func validOwner(value string) bool {
	return value == "iam" || value == "forms" || value == "workflow" ||
		value == "analytics" || value == "catalog"
}

func (value Plan) Valid() bool {
	if !value.PackageID.Valid() || !value.Version.Valid() ||
		!slugPattern.MatchString(value.Slug) || !digestPattern.MatchString(value.PackageSHA256) ||
		!digestPattern.MatchString(value.SHA256) || !validPins(value) {
		return false
	}
	rebuilt, err := NewSealedPlan(value.PackageID, value.Version, value.Slug,
		value.PackageSHA256, value.Dependencies, value.Components, value.Menu)
	return err == nil && rebuilt.SHA256 == value.SHA256
}

func (value Plan) Clone() Plan {
	value.Dependencies = append([]DependencyPin(nil), value.Dependencies...)
	value.Components = append([]ComponentPin(nil), value.Components...)
	value.Menu = append([]MenuPin(nil), value.Menu...)
	for index := range value.Menu {
		if value.Menu[index].ParentID != nil {
			parent := *value.Menu[index].ParentID
			value.Menu[index].ParentID = &parent
		}
		if value.Menu[index].TargetID != nil {
			target := *value.Menu[index].TargetID
			value.Menu[index].TargetID = &target
		}
	}
	return value
}
