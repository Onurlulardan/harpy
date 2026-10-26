package rbac

type Implication struct {
	ID                    ID
	SourcePermissionID    ID
	ImpliedPermissionID   ID
	RetirementContractKey string
	State                 string
	HitCount              int64
	Revision              int64
}

type ImplicationInput struct {
	ID                    ID
	SourcePermissionID    ID
	ImpliedPermissionID   ID
	RetirementContractKey string
	Evidence              MutationEvidence
}

type RecordHitCommand struct {
	ID               ID
	ExpectedRevision int64
	Evidence         MutationEvidence
}
