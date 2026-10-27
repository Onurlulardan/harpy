package iam

import (
	"context"
	"fmt"
	"time"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	platformiampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/platformiam"
	securityepochpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/securityepoch"
	sessionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/sessionview"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformiam "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/platformiam"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/securityepoch"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/sessionview"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type BillingAccessPlatform interface {
	IdentityHTTPPlatform
	ResolveActiveTenantIDBySlug(context.Context, string) (uuid.UUID, error)
	ResolveGlobalBillingGenerations(context.Context) (int64, int64, error)
}

type BillingAccessRuntime struct {
	pool        BillingAccessPlatform
	coordinator *platformtx.Coordinator
	security    *HTTPSecurity
	authorizer  authorization.Authorizer
	platformIAM *platformiam.Service
	epoch       *securityepoch.Asserter
	sessions    *sessionview.Reader
	clock       func() time.Time
}

func NewBillingAccessRuntime(
	pool BillingAccessPlatform,
	coordinator *platformtx.Coordinator,
	security *HTTPSecurity,
	clock func() time.Time,
) (*BillingAccessRuntime, error) {
	if pool == nil || coordinator == nil || !security.Valid() || clock == nil {
		return nil, fmt.Errorf("IAM Billing access dependencies are required")
	}
	port, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM Billing authorization evidence: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return nil, fmt.Errorf("build IAM Billing authorizer: %w", err)
	}
	platformStore, err := platformiampostgres.NewStore()
	if err != nil {
		return nil, fmt.Errorf("build IAM Billing platform authority: %w", err)
	}
	platformAuthority, err := platformiam.NewService(platformStore, iamPlatformClock{})
	if err != nil {
		return nil, fmt.Errorf("build IAM Billing platform service: %w", err)
	}
	epochStore, err := securityepochpostgres.NewStore()
	if err != nil {
		return nil, fmt.Errorf("build IAM Billing security epoch store: %w", err)
	}
	epochAsserter, err := securityepoch.NewAsserter(epochStore)
	if err != nil {
		return nil, fmt.Errorf("build IAM Billing security epoch asserter: %w", err)
	}
	sessionStore, err := sessionpostgres.NewStore()
	if err != nil {
		return nil, fmt.Errorf("build IAM Billing session projection store: %w", err)
	}
	sessions, err := sessionview.NewReader(sessionStore)
	if err != nil {
		return nil, fmt.Errorf("build IAM Billing session projection reader: %w", err)
	}
	return &BillingAccessRuntime{
		pool: pool, coordinator: coordinator, security: security,
		authorizer: authorizer, platformIAM: platformAuthority,
		epoch: epochAsserter, sessions: sessions, clock: clock,
	}, nil
}
