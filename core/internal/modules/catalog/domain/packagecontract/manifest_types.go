package packagecontract

type Profile string
type DependencyKind string
type ComponentKind string
type MenuKind string

const (
	ProfileCore     Profile = "core"
	ProfileBusiness Profile = "business"

	DependencyRequired DependencyKind = "required"
	DependencyOptional DependencyKind = "optional"

	ComponentForm      ComponentKind = "FORM"
	ComponentWorkflow  ComponentKind = "WORKFLOW"
	ComponentDashboard ComponentKind = "DASHBOARD"
	ComponentEndpoint  ComponentKind = "ENDPOINT"
	ComponentIAM       ComponentKind = "IAM_TEMPLATE"
	ComponentDemo      ComponentKind = "DEMO_DATA"

	MenuGroup        MenuKind = "GROUP"
	MenuForm         MenuKind = "FORM"
	MenuWorkflow     MenuKind = "WORKFLOW"
	MenuDashboard    MenuKind = "DASHBOARD"
	MenuEndpoint     MenuKind = "ENDPOINT"
	MenuExternalLink MenuKind = "EXTERNAL_LINK"
)

type Dependency struct {
	PackageID PackageID
	Range     Range
	Kind      DependencyKind
}

type Component struct {
	ID          ComponentID
	Kind        ComponentKind
	Owner       string
	Contract    string
	ContentHash string
}

type MenuItem struct {
	ID       MenuItemID
	ParentID *MenuItemID
	Kind     MenuKind
	TargetID *ComponentID
	Label    string
	Order    int
	URL      string
}

type Manifest struct {
	PackageID        PackageID
	Slug             string
	Version          Version
	PlatformContract string
	Profile          Profile
	Dependencies     []Dependency
	Components       []Component
	Menu             []MenuItem
	canonical        []byte
}

func (value Manifest) Canonical() []byte { return append([]byte(nil), value.canonical...) }
