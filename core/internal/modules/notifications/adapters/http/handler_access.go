package notificationhttp

import "net/http"

func (handler *Handler) accessFor(r *http.Request, permission Permission,
	operation string,
) (Access, error) {
	if handler == nil || r == nil || nilValue(handler.access) ||
		!validPermissionOperation(permission, operation) {
		return Access{}, ErrAuthentication
	}
	session, csrf, origin := operationSecurity(operation)
	access, err := handler.access.Verify(r.Context(), AccessRequest{
		AuthorizationValues: append([]string(nil), r.Header.Values("Authorization")...),
		CookieValues:        append([]string(nil), r.Header.Values("Cookie")...),
		TenantIDValues:      append([]string(nil), r.Header.Values("X-Tenant-ID")...),
		TenantSlugValues:    append([]string(nil), r.Header.Values("X-Tenant-Slug")...),
		CSRFValues:          append([]string(nil), r.Header.Values("X-CSRF-Token")...),
		OriginValues:        append([]string(nil), r.Header.Values("Origin")...),
		ConsumerManifestValues: append([]string(nil),
			r.Header.Values("X-Harpy-Consumer-Manifest")...),
		ConsumerSignatureValues: append([]string(nil),
			r.Header.Values("X-Harpy-Consumer-Signature")...),
		Permission: permission, Operation: operation, RequireSession: session,
		RequireCSRF: csrf, RequireHTTPSOrigin: origin, RequireSecureTransport: origin,
		RequireSignedConsumer: operation == OperationLegacyConnect,
	})
	if err != nil {
		return Access{}, err
	}
	if !access.ValidFor(permission, operation) {
		return Access{}, ErrForbidden
	}
	return access, nil
}
