package publication

type Audience string

const (
	AudienceTenant     Audience = "tenant"
	AudiencePublic     Audience = "public"
	AudienceCapability Audience = "capability"
)

func (value Audience) Valid() bool {
	switch value {
	case AudienceTenant, AudiencePublic, AudienceCapability:
		return true
	default:
		return false
	}
}

type State string

const (
	StateDraft           State = "draft"
	StatePendingApproval State = "pending_approval"
	StateActive          State = "active"
	StateRevoked         State = "revoked"
	StateRetired         State = "retired"
)

func (value State) Valid() bool {
	switch value {
	case StateDraft, StatePendingApproval, StateActive, StateRevoked, StateRetired:
		return true
	default:
		return false
	}
}

type CachePolicy string

const (
	CacheNone          CachePolicy = "none"
	CacheRevisionBound CachePolicy = "revision_bound"
)

func (value CachePolicy) Valid() bool {
	return value == CacheNone || value == CacheRevisionBound
}
