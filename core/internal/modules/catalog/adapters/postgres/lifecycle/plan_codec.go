package lifecycle

import (
	"encoding/json"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

type planDocument struct {
	PackageID     string               `json:"package_id"`
	Slug          string               `json:"slug"`
	PackageSHA256 string               `json:"package_sha256"`
	Version       string               `json:"version"`
	Dependencies  []dependencyDocument `json:"dependencies"`
	Components    []componentDocument  `json:"components"`
	Menu          []menuDocument       `json:"menu"`
}

type dependencyDocument struct {
	PackageID string `json:"package_id"`
	Version   string `json:"version"`
	Required  bool   `json:"required"`
}

type componentDocument struct {
	ComponentID     string `json:"component_id"`
	Owner           string `json:"owner"`
	OwnerGeneration int64  `json:"owner_generation"`
	Contract        string `json:"contract"`
	ContentSHA256   string `json:"content_sha256"`
}

type menuDocument struct {
	ItemID   string `json:"item_id"`
	ParentID string `json:"parent_id"`
	Kind     string `json:"kind"`
	TargetID string `json:"target_id"`
	Label    string `json:"label"`
	Order    int    `json:"order"`
	URL      string `json:"url"`
}

func encodePlan(value catalogdomain.Plan) ([]byte, error) {
	if !value.Valid() {
		return nil, unavailable()
	}
	document := planDocument{PackageID: value.PackageID.String(), Slug: value.Slug,
		PackageSHA256: value.PackageSHA256, Version: value.Version.String(),
		Dependencies: make([]dependencyDocument, 0, len(value.Dependencies)),
		Components:   make([]componentDocument, 0, len(value.Components)),
		Menu:         make([]menuDocument, 0, len(value.Menu))}
	for _, pin := range value.Dependencies {
		document.Dependencies = append(document.Dependencies,
			dependencyDocument{PackageID: pin.PackageID.String(), Version: pin.Version.String(),
				Required: pin.Required})
	}
	for _, pin := range value.Components {
		document.Components = append(document.Components, componentDocument{ComponentID: pin.ComponentID.String(),
			Owner: pin.Owner, OwnerGeneration: pin.OwnerGeneration,
			Contract: pin.Contract, ContentSHA256: pin.ContentSHA256})
	}
	for _, pin := range value.Menu {
		item := menuDocument{ItemID: pin.ItemID.String(), Kind: string(pin.Kind),
			Label: pin.Label, Order: pin.Order, URL: pin.URL}
		if pin.ParentID != nil {
			item.ParentID = pin.ParentID.String()
		}
		if pin.TargetID != nil {
			item.TargetID = pin.TargetID.String()
		}
		document.Menu = append(document.Menu, item)
	}
	return json.Marshal(document)
}

func decodePlan(raw []byte, expectedSHA256 string) (catalogdomain.Plan, error) {
	var document planDocument
	if len(raw) == 0 || json.Unmarshal(raw, &document) != nil {
		return catalogdomain.Plan{}, unavailable()
	}
	packageID, err := packagecontract.ParsePackageID(document.PackageID)
	version, versionErr := packagecontract.ParseVersion(document.Version)
	if err != nil || versionErr != nil {
		return catalogdomain.Plan{}, unavailable()
	}
	dependencies := make([]catalogdomain.DependencyPin, 0, len(document.Dependencies))
	for _, item := range document.Dependencies {
		id, parseErr := packagecontract.ParsePackageID(item.PackageID)
		pinVersion, pinErr := packagecontract.ParseVersion(item.Version)
		if parseErr != nil || pinErr != nil {
			return catalogdomain.Plan{}, unavailable()
		}
		dependencies = append(dependencies, catalogdomain.DependencyPin{
			PackageID: id, Version: pinVersion, Required: item.Required})
	}
	components := make([]catalogdomain.ComponentPin, 0, len(document.Components))
	for _, item := range document.Components {
		id, parseErr := packagecontract.ParseComponentID(item.ComponentID)
		if parseErr != nil {
			return catalogdomain.Plan{}, unavailable()
		}
		components = append(components, catalogdomain.ComponentPin{ComponentID: id, Owner: item.Owner,
			OwnerGeneration: item.OwnerGeneration, Contract: item.Contract, ContentSHA256: item.ContentSHA256})
	}
	menu := make([]catalogdomain.MenuPin, 0, len(document.Menu))
	for _, item := range document.Menu {
		pin, parseErr := decodeMenuPin(item)
		if parseErr != nil {
			return catalogdomain.Plan{}, unavailable()
		}
		menu = append(menu, pin)
	}
	plan, err := catalogdomain.NewSealedPlan(packageID, version, document.Slug,
		document.PackageSHA256, dependencies, components, menu)
	if err != nil || plan.SHA256 != expectedSHA256 {
		return catalogdomain.Plan{}, unavailable()
	}
	return plan, nil
}

func decodeMenuPin(item menuDocument) (catalogdomain.MenuPin, error) {
	id, err := packagecontract.ParseMenuItemID(item.ItemID)
	value := catalogdomain.MenuPin{ItemID: id, Kind: packagecontract.MenuKind(item.Kind),
		Label: item.Label, Order: item.Order, URL: item.URL}
	if err != nil {
		return catalogdomain.MenuPin{}, err
	}
	if item.ParentID != "" {
		parent, parseErr := packagecontract.ParseMenuItemID(item.ParentID)
		if parseErr != nil {
			return catalogdomain.MenuPin{}, parseErr
		}
		value.ParentID = &parent
	}
	if item.TargetID != "" {
		target, parseErr := packagecontract.ParseComponentID(item.TargetID)
		if parseErr != nil {
			return catalogdomain.MenuPin{}, parseErr
		}
		value.TargetID = &target
	}
	return value, nil
}
