package billingapi

type MeteredResourceKind string

const (
	MeteredResourceAPIEndpoint        MeteredResourceKind = "api_endpoint"
	MeteredResourceAnalyticsDashboard MeteredResourceKind = "analytics_dashboard"
)

func (kind MeteredResourceKind) Valid() bool {
	return kind == MeteredResourceAPIEndpoint ||
		kind == MeteredResourceAnalyticsDashboard
}

type MeteredResourceMutation string

const (
	MeteredResourceCreate     MeteredResourceMutation = "create"
	MeteredResourceUpdate     MeteredResourceMutation = "update"
	MeteredResourcePublish    MeteredResourceMutation = "publish"
	MeteredResourceRevalidate MeteredResourceMutation = "revalidate"
	MeteredResourceDelete     MeteredResourceMutation = "delete"
	MeteredResourceRestore    MeteredResourceMutation = "restore"
)

func (mutation MeteredResourceMutation) Valid() bool {
	switch mutation {
	case MeteredResourceCreate, MeteredResourceUpdate, MeteredResourcePublish,
		MeteredResourceRevalidate, MeteredResourceDelete, MeteredResourceRestore:
		return true
	default:
		return false
	}
}

func (mutation MeteredResourceMutation) RequiresBuilder() bool {
	return mutation != MeteredResourceDelete && mutation.Valid()
}

func (mutation MeteredResourceMutation) GaugeDelta() int64 {
	switch mutation {
	case MeteredResourceCreate, MeteredResourceRestore:
		return 1
	case MeteredResourceDelete:
		return -1
	default:
		return 0
	}
}
