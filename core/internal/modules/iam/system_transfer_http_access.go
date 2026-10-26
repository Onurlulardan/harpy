package iam

import "net/http"

func (handler *SystemTransferHTTP) transferAccess(r *http.Request,
	resource SystemTransferResource, action, operation string) (SystemTransferAccess, error) {
	if handler == nil || systemTransferNil(handler.access) || r == nil {
		return SystemTransferAccess{}, ErrSystemTransferAuthentication
	}
	permission := resource.Permission(action)
	if permission == "" {
		return SystemTransferAccess{}, ErrSystemTransferForbidden
	}
	access, err := handler.access.VerifySystemTransfer(r.Context(),
		SystemTransferAccessRequest{
			AuthorizationValues: append([]string(nil), r.Header.Values("Authorization")...),
			TenantIDValues:      append([]string(nil), r.Header.Values("X-Tenant-ID")...),
			TenantSlugValues:    append([]string(nil), r.Header.Values("X-Tenant-Slug")...),
			Resource:            resource, Permission: permission, Operation: operation,
		})
	if err != nil {
		return SystemTransferAccess{}, err
	}
	if !access.ValidFor(resource, permission, operation) {
		return SystemTransferAccess{}, ErrSystemTransferForbidden
	}
	return access, nil
}
