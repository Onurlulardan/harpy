package schemav1

type Plan struct {
	ContractVersion   string        `json:"contract_version"`
	TenantID          string        `json:"tenant_id"`
	Owner             string        `json:"owner"`
	Resource          Resource      `json:"resource"`
	Compiler          Compiler      `json:"compiler"`
	AllowedOperations []OperationID `json:"allowed_operations"`
	Steps             []Step        `json:"steps"`
}

type Step struct {
	Ordinal           int                `json:"ordinal"`
	OperationID       OperationID        `json:"operation_id"`
	TransactionClass  TransactionClass   `json:"transaction_class"`
	ExecutorKind      ExecutorKind       `json:"executor_kind"`
	IdentifierMapping IdentifierMapping  `json:"identifier_mapping"`
	Arguments         OperationArguments `json:"arguments"`
}
