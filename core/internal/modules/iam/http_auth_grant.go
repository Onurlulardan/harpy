package iam

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/csrf"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/refreshtoken"
)

type authHTTPGrantResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (handler *AuthHTTP) prepareGrant(
	ctx context.Context,
	grant AuthHTTPGrant,
	origin string,
	now time.Time,
) (authHTTPGrantResponse, []string, error) {
	if nilAuthHTTPGrant(grant) || !handler.security.browserOriginAllowed(origin) {
		return authHTTPGrantResponse{}, nil, fmt.Errorf("IAM auth grant is invalid")
	}
	access, err := grant.AccessToken()
	if err != nil {
		return authHTTPGrantResponse{}, nil, fmt.Errorf("IAM access grant is invalid")
	}
	claims, err := handler.security.accessVerifier.Verify(ctx, access, now)
	if err != nil {
		return authHTTPGrantResponse{}, nil, fmt.Errorf("IAM access grant is unverifiable")
	}
	expiresAt := time.Unix(claims.ExpiresAtUnix(), 0).UTC()
	absoluteExpiry := grant.AbsoluteExpiresAt().UTC()
	if !absoluteExpiry.After(now) || expiresAt.Before(now) {
		return authHTTPGrantResponse{}, nil, fmt.Errorf("IAM auth grant is expired")
	}
	cookies, err := handler.revealGrantCookies(
		grant, origin, absoluteExpiry, now,
	)
	if err != nil {
		return authHTTPGrantResponse{}, nil, err
	}
	return authHTTPGrantResponse{Token: access, ExpiresAt: expiresAt}, cookies, nil
}

func (handler *AuthHTTP) revealGrantCookies(
	grant AuthHTTPGrant,
	origin string,
	absoluteExpiry time.Time,
	now time.Time,
) ([]string, error) {
	var cookies []string
	err := grant.RevealRefreshOnce(func(wire []byte) error {
		presented, parseErr := refreshtoken.Parse(wire)
		if parseErr != nil {
			return fmt.Errorf("parse issued refresh credential")
		}
		defer presented.Destroy()
		selector, selectorErr := presented.Selector()
		if selectorErr != nil {
			return fmt.Errorf("read issued refresh selector")
		}
		refreshValue, valid := refreshCookieValue(wire, absoluteExpiry, now)
		if !valid {
			return fmt.Errorf("build issued refresh cookie")
		}
		csrfCredential, signErr := handler.security.csrfSigner.Sign(selector, origin, now)
		if signErr != nil {
			return fmt.Errorf("sign issued CSRF credential")
		}
		defer csrfCredential.Destroy()
		var csrfValue string
		revealErr := csrfCredential.RevealOnce(func(csrfWire []byte) error {
			var ok bool
			csrfValue, ok = csrfCookieValue(csrfWire, now)
			if !ok {
				return csrf.ErrInvalidCredential
			}
			return nil
		})
		if revealErr != nil {
			return fmt.Errorf("build issued CSRF cookie")
		}
		cookies = []string{refreshValue, csrfValue}
		return nil
	})
	if err != nil || len(cookies) != 2 {
		return nil, fmt.Errorf("reveal IAM browser grant")
	}
	return cookies, nil
}

func nilAuthHTTPGrant(grant AuthHTTPGrant) bool {
	if grant == nil {
		return true
	}
	value := reflect.ValueOf(grant)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
