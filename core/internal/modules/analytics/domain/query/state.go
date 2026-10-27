package query

type Lifecycle string

const (
	LifecycleActive  Lifecycle = "active"
	LifecycleRetired Lifecycle = "retired"
)

func (state Lifecycle) Valid() bool {
	return state == LifecycleActive || state == LifecycleRetired
}

type VersionState string

const (
	VersionDraft             VersionState = "draft"
	VersionPublished         VersionState = "published"
	VersionNeedsRevalidation VersionState = "needs_revalidation"
	VersionRetired           VersionState = "retired"
)

func (state VersionState) Valid() bool {
	switch state {
	case VersionDraft, VersionPublished, VersionNeedsRevalidation, VersionRetired:
		return true
	default:
		return false
	}
}
