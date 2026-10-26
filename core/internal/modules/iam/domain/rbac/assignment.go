package rbac

import "time"

type AssignmentTarget string

const (
	AssignmentUser         AssignmentTarget = "USER"
	AssignmentRole         AssignmentTarget = "ROLE"
	AssignmentOrganization AssignmentTarget = "ORGANIZATION"
)

type AssignmentItem struct {
	Ordinal      int
	ActionID     ID
	PermissionID ID
	BindingID    ID
}

type Assignment struct {
	ID         ID
	ResourceID ID
	Target     AssignmentTarget
	TargetID   ID
	State      string
	Revision   int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Items      []AssignmentItem
}

type StepUpEvidence struct {
	PurposeSHA256 [32]byte
	AuthTime      time.Time
	AMRPassword   bool
	AMRWebAuthn   bool
	AMRTOTP       bool
}

type AssignmentMutation struct {
	ID               ID
	ResourceID       ID
	Target           AssignmentTarget
	TargetID         ID
	Items            []AssignmentItem
	ExpectedRevision int64
	Evidence         MutationEvidence
	StepUp           StepUpEvidence
}

type AssignmentGetQuery struct {
	ID       ID
	Evidence QueryEvidence
}

type AssignmentListQuery struct {
	State    string
	Limit    int
	AfterID  ID
	Evidence QueryEvidence
}

type EffectivePermission struct {
	PermissionID ID
	MatchKind    string
	ResourceID   ID
	ResourceKey  string
	ActionID     ID
	ActionKey    string
	Source       AssignmentTarget
	SourceID     ID
}

type EffectivePermissionQuery struct {
	Target   AssignmentTarget
	TargetID ID
	Limit    int
	Evidence QueryEvidence
}

func (target AssignmentTarget) Valid() bool {
	return target == AssignmentUser || target == AssignmentRole ||
		target == AssignmentOrganization
}

func (evidence StepUpEvidence) Validate() error {
	if evidence.PurposeSHA256 == [32]byte{} || evidence.AuthTime.IsZero() ||
		!(evidence.AMRPassword || evidence.AMRWebAuthn || evidence.AMRTOTP) {
		return ErrInvalidMutation
	}
	return nil
}

func (command AssignmentMutation) ValidateCreate() error {
	return command.validate(0)
}

func (command AssignmentMutation) ValidateUpdate() error {
	if command.ExpectedRevision < 1 {
		return ErrInvalidMutation
	}
	return command.validate(command.ExpectedRevision)
}

func (command AssignmentMutation) ValidateRetire() error {
	if !command.ID.Valid() || command.ExpectedRevision < 1 ||
		command.Evidence.Validate() != nil || command.StepUp.Validate() != nil {
		return ErrInvalidMutation
	}
	return nil
}

func (command AssignmentMutation) validate(expected int64) error {
	if !command.ID.Valid() || !command.ResourceID.Valid() || !command.Target.Valid() ||
		!command.TargetID.Valid() || command.ExpectedRevision != expected ||
		len(command.Items) < 1 || len(command.Items) > 100 ||
		command.Evidence.Validate() != nil || command.StepUp.Validate() != nil {
		return ErrInvalidMutation
	}
	seenActions := make(map[ID]struct{}, len(command.Items))
	seenPermissions := make(map[ID]struct{}, len(command.Items))
	seenBindings := make(map[ID]struct{}, len(command.Items))
	for index, item := range command.Items {
		_, actionSeen := seenActions[item.ActionID]
		_, permissionSeen := seenPermissions[item.PermissionID]
		_, bindingSeen := seenBindings[item.BindingID]
		if item.Ordinal != index+1 || !item.ActionID.Valid() ||
			!item.PermissionID.Valid() || !item.BindingID.Valid() ||
			actionSeen || permissionSeen || bindingSeen {
			return ErrInvalidMutation
		}
		seenActions[item.ActionID] = struct{}{}
		seenPermissions[item.PermissionID] = struct{}{}
		seenBindings[item.BindingID] = struct{}{}
	}
	return nil
}

func (query AssignmentGetQuery) Validate() error {
	if !query.ID.Valid() || query.Evidence.Validate() != nil {
		return ErrInvalidQuery
	}
	return nil
}

func (query AssignmentListQuery) Validate() error {
	if query.Limit < 1 || query.Limit > 200 ||
		(query.State != "active" && query.State != "retired" && query.State != "all") ||
		query.Evidence.Validate() != nil {
		return ErrInvalidQuery
	}
	return nil
}

func (query EffectivePermissionQuery) Validate() error {
	if !query.Target.Valid() || !query.TargetID.Valid() || query.Limit < 1 ||
		query.Limit > 500 || query.Evidence.Validate() != nil {
		return ErrInvalidQuery
	}
	return nil
}
