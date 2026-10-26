package httpapi

type AccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          Permission
	Operation           string
}
