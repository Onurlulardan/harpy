package fieldregistry

type StorageLayout string
type StorageType string
type StorageFeature string
type QueryType string
type Gate string
type Operator string
type AggregateFunction string
type CostClass string
type BudgetDimension string
type ValueProfile string

type OperatorCapability struct {
	Operator  Operator
	Gate      Gate
	CountSafe bool
}

type AggregateCapability struct {
	Function AggregateFunction
	Gate     Gate
}
