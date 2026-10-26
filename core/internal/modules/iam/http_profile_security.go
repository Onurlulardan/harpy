package iam

import (
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (handler *ProfileHTTP) claims(
	r *http.Request,
) (accesstoken.Claims, bool) {
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
	if _, portal := claims.PortalBinding(); portal {
		return accesstoken.Claims{}, false
	}
	if values := r.Header.Values(authTenantSlugHeader); len(values) > 0 {
		if _, valid := exactNonempty(values, 128); !valid {
			return accesstoken.Claims{}, false
		}
	}
	return claims, true
}
