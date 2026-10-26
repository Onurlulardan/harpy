package fieldregistry

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
