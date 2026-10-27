package iam

import (
	"context"
	"fmt"
	"time"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	portaluserpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/portaluser"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type FormsRecordAccessPlatform interface {
	IdentityHTTPPlatform
	ResolveActiveTenantIDBySlug(context.Context, string) (uuid.UUID, error)
}

type FormsRecordAccessRuntime struct {
	pool        FormsRecordAccessPlatform
	coordinator *platformtx.Coordinator
	security    *HTTPSecurity
	authorizer  authorization.Authorizer
	portalUsers *portaluser.Service
	clock       func() time.Time
}

func NewFormsRecordAccessRuntime(
	pool FormsRecordAccessPlatform,
	coordinator *platformtx.Coordinator,
	security *HTTPSecurity,
	clock func() time.Time,
) (*FormsRecordAccessRuntime, error) {
	if pool == nil || coordinator == nil || !security.Valid() || clock == nil {
		return nil, fmt.Errorf("IAM Forms record access dependencies are required")
	}
	port, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM Forms record authorization evidence: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return nil, fmt.Errorf("build IAM Forms record authorizer: %w", err)
	}
	portalStore, err := portaluserpostgres.NewStore()
	if err != nil {
		return nil, fmt.Errorf("build IAM Forms portal projection storage: %w", err)
	}
	portalUsers, err := portaluser.NewService(portalStore)
	if err != nil {
		return nil, fmt.Errorf("build IAM Forms portal projection service: %w", err)
	}
	return &FormsRecordAccessRuntime{
		pool: pool, coordinator: coordinator, security: security,
		authorizer: authorizer, portalUsers: portalUsers, clock: clock,
	}, nil
}
