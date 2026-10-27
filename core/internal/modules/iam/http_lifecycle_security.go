package iam

import (
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (handler *LifecycleHTTP) claims(r *http.Request) (accesstoken.Claims, bool) {
	if handler == nil || r == nil {
		return accesstoken.Claims{}, false
	}
	now := handler.clock().UTC().Truncate(time.Second)
	wire, valid := exactBearer(r)
	if !valid || now.IsZero() || now.Unix() <= 0 {
		return accesstoken.Claims{}, false
	}
	claims, err := handler.security.accessVerifier.Verify(r.Context(), wire, now)
	if err != nil || !optionalTenantIDMatches(r.Header.Values(authTenantIDHeader), claims.Tenant()) {
		return accesstoken.Claims{}, false
	}
	switch claims.Client() {
	case accesstoken.ClientOperatorConsole:
		if claims.AccessMode() != accesstoken.AccessModeOperator {
			return accesstoken.Claims{}, false
		}
	case accesstoken.ClientPortal:
		if claims.AccessMode() != accesstoken.AccessModePortal {
			return accesstoken.Claims{}, false
		}
	default:
		return accesstoken.Claims{}, false
	}
	return claims, true
}
