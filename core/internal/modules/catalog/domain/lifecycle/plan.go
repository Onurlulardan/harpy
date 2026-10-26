package lifecycle

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

type DependencyPin struct {
	PackageID packagecontract.PackageID
	Version   packagecontract.Version
	Required  bool
}

type ComponentPin struct {
	ComponentID     packagecontract.ComponentID
	Owner           string
	OwnerGeneration int64
	Contract        string
	ContentSHA256   string
}

type MenuPin struct {
	ItemID   packagecontract.MenuItemID
	ParentID *packagecontract.MenuItemID
	Kind     packagecontract.MenuKind
	TargetID *packagecontract.ComponentID
	Label    string
	Order    int
	URL      string
}

type Plan struct {
	PackageID     packagecontract.PackageID
	Slug          string
	PackageSHA256 string
	Version       packagecontract.Version
	Dependencies  []DependencyPin
	Components    []ComponentPin
	Menu          []MenuPin
	SHA256        string
}

func NewPlan(
	packageID packagecontract.PackageID,
	version packagecontract.Version,
	dependencies []DependencyPin,
	components []ComponentPin,
) (Plan, error) {
	return NewPlanWithMenu(packageID, version, dependencies, components, nil)
}

func NewPlanWithMenu(
	packageID packagecontract.PackageID,
	version packagecontract.Version,
	dependencies []DependencyPin,
	components []ComponentPin,
	menu []MenuPin,
) (Plan, error) {
	digest := sha256.Sum256([]byte(packageID.String() + "\x00" + version.String()))
	return NewSealedPlan(packageID, version, "package-"+packageID.String()[:8],
		hex.EncodeToString(digest[:]), dependencies, components, menu)
}

func NewSealedPlan(
	packageID packagecontract.PackageID,
	version packagecontract.Version,
	slug, packageSHA256 string,
	dependencies []DependencyPin,
	components []ComponentPin,
	menu []MenuPin,
) (Plan, error) {
	if !packageID.Valid() || !version.Valid() || len(components) == 0 ||
		len(components) > 256 || len(dependencies) > 64 || len(menu) > 512 {
		return Plan{}, ErrInvalid
	}
	value := Plan{PackageID: packageID, Slug: slug, PackageSHA256: packageSHA256, Version: version,
		Dependencies: append([]DependencyPin(nil), dependencies...),
		Components:   append([]ComponentPin(nil), components...),
		Menu:         append([]MenuPin(nil), menu...)}
	if !validPins(value) {
		return Plan{}, ErrInvalid
	}
	value.sort()
	digest := sha256.Sum256(value.digestDocument())
	value.SHA256 = hex.EncodeToString(digest[:])
	return value, nil
}

func (value *Plan) sort() {
	sort.Slice(value.Dependencies, func(i, j int) bool {
		return value.Dependencies[i].PackageID.String() < value.Dependencies[j].PackageID.String()
	})
	sort.Slice(value.Components, func(i, j int) bool {
		left, right := value.Components[i], value.Components[j]
		if left.Owner != right.Owner {
			return left.Owner < right.Owner
		}
		return left.ComponentID.String() < right.ComponentID.String()
	})
	sort.Slice(value.Menu, func(i, j int) bool {
		if value.Menu[i].Order != value.Menu[j].Order {
			return value.Menu[i].Order < value.Menu[j].Order
		}
		return value.Menu[i].ItemID.String() < value.Menu[j].ItemID.String()
	})
}

func (value Plan) digestDocument() []byte {
	var buffer bytes.Buffer
	writePlanField(&buffer, "catalog-operation-plan.v1")
	writePlanField(&buffer, value.PackageID.String())
	writePlanField(&buffer, value.Slug)
	writePlanField(&buffer, value.PackageSHA256)
	writePlanField(&buffer, value.Version.String())
	for _, pin := range value.Dependencies {
		writePlanField(&buffer, pin.PackageID.String())
		writePlanField(&buffer, pin.Version.String())
		if pin.Required {
			writePlanField(&buffer, "required")
		} else {
			writePlanField(&buffer, "optional")
		}
	}
	for _, pin := range value.Components {
		writePlanField(&buffer, pin.Owner)
		writePlanField(&buffer, pin.ComponentID.String())
		_ = binary.Write(&buffer, binary.BigEndian, pin.OwnerGeneration)
		writePlanField(&buffer, pin.Contract)
		writePlanField(&buffer, pin.ContentSHA256)
	}
	for _, pin := range value.Menu {
		writePlanField(&buffer, pin.ItemID.String())
		if pin.ParentID != nil {
			writePlanField(&buffer, pin.ParentID.String())
		} else {
			writePlanField(&buffer, "")
		}
		writePlanField(&buffer, string(pin.Kind))
		if pin.TargetID != nil {
			writePlanField(&buffer, pin.TargetID.String())
		} else {
			writePlanField(&buffer, "")
		}
		writePlanField(&buffer, pin.Label)
		_ = binary.Write(&buffer, binary.BigEndian, int64(pin.Order))
		writePlanField(&buffer, pin.URL)
	}
	return buffer.Bytes()
}

func writePlanField(buffer *bytes.Buffer, value string) {
	_ = binary.Write(buffer, binary.BigEndian, uint32(len(value)))
	buffer.WriteString(value)
}
