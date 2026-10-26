package idempotency

type RetentionClass string

const (
	InteractiveWrite    RetentionClass = "interactive_write"
	AdministrativeWrite RetentionClass = "administrative_write"
	AsyncHandle         RetentionClass = "async_handle"
)

func (class RetentionClass) valid() bool {
	return class == InteractiveWrite || class == AdministrativeWrite || class == AsyncHandle
}
