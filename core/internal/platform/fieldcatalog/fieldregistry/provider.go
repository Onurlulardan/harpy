package fieldregistry

const ContractVersion = "forms-field-registry.v1"

type Snapshot struct {
	ContractVersion string
	HardLimits      BudgetLimits
	Types           []Descriptor
}

type Provider interface {
	Resolve(TypeID) (Descriptor, error)
	ValidateRole(TypeID, Role) (Descriptor, error)
	Snapshot() (Snapshot, error)
}
