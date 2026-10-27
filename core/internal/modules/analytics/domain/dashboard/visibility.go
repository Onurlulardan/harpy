package dashboard

type Visibility string

const (
	VisibilityPrivate Visibility = "private"
	VisibilityTenant  Visibility = "tenant"
)

func (value Visibility) Valid() bool {
	return value == VisibilityPrivate || value == VisibilityTenant
}

type Lifecycle string

const (
	LifecycleActive  Lifecycle = "active"
	LifecycleRetired Lifecycle = "retired"
)

func (value Lifecycle) Valid() bool {
	return value == LifecycleActive || value == LifecycleRetired
}
