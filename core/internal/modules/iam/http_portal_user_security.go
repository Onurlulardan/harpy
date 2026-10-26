package iam

import (
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (handler *PortalUserHTTP) claims(r *http.Request) (accesstoken.Claims, bool) {
	if handler == nil || r == nil {
		return accesstoken.Claims{}, false
	}
	now := handler.clock().UTC().Truncate(time.Second)
	wire, valid := exactBearer(r)
	if !valid || now.IsZero() || now.Unix() <= 0 {
		return accesstoken.Claims{}, false
	}
	claims, err := handler.security.accessVerifier.Verify(r.Context(), wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!optionalTenantIDMatches(r.Header.Values(authTenantIDHeader), claims.Tenant()) {
		return accesstoken.Claims{}, false
	}
	_, portal := claims.PortalBinding()
	return claims, !portal
}
