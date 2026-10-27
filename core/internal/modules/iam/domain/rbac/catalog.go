package rbac

type Resource struct {
	ID          ID
	Key         string
	OwnerDomain string
	Kind        string
	DisplayName string
	Description string
	Protected   bool
	State       string
	Revision    int64
}

type Action struct {
	ID          ID
	Key         string
	DisplayName string
	Description string
	Protected   bool
	State       string
	Revision    int64
}

type Permission struct {
	ID         ID
	MatchKind  string
	ResourceID ID
	ActionID   ID
	Protected  bool
	State      string
	Revision   int64
}

type ResourceInput struct {
	ID          ID
	Key         string
	OwnerDomain string
	Kind        string
	DisplayName string
	Description string
	Protected   bool
	Evidence    MutationEvidence
}

type ActionInput struct {
	ID          ID
	Key         string
	DisplayName string
	Description string
	Protected   bool
	Evidence    MutationEvidence
}

type PermissionInput struct {
	ID         ID
	MatchKind  string
	ResourceID ID
	ActionID   ID
	Protected  bool
	Evidence   MutationEvidence
}
