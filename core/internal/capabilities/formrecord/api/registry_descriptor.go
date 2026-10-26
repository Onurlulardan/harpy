package recordkernel

type OperatorCapability struct {
	Operator  Operator
	Gate      Gate
	CountSafe bool
}

type AggregateCapability struct {
	Function AggregateFunction
	Gate     Gate
}

type StorageCapability struct {
	Type            StorageType
	Layout          StorageLayout
	AllowedFeatures []StorageFeature
}

type QueryCapability struct {
	LogicalType         QueryType
	SelectGate          Gate
	SortGate            Gate
	SearchGate          Gate
	GroupGate           Gate
	RequiresCursorBound bool
	Operators           []OperatorCapability
	Aggregates          []AggregateCapability
}

type CostProfile struct {
	Class      CostClass
	Dimensions []BudgetDimension
}

type Descriptor struct {
	TypeID       TypeID
	ValueProfile ValueProfile
	Roles        []Role
	Storage      StorageCapability
	Query        QueryCapability
	Cost         CostProfile
}

type RegistrySnapshot struct {
	ContractVersion string
	HardLimits      BudgetLimits
	Types           []Descriptor
}
