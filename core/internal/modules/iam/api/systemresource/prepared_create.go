package systemresource

const PreparedCreateContract = "iam-system-resource-prepared-create.v1"

// PreparedCreate is an opaque, one-shot IAM-owned create preparation. The
// concrete implementation remains private to IAM so password work can happen
// before a database transaction is opened.
type PreparedCreate interface {
	Resource() (Resource, error)
	Contract() string
	Destroy()
}
