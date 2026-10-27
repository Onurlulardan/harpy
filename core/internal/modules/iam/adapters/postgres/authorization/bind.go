package postgres

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var errAuthorizationRead = errors.New("IAM authorization evidence read failed")

type tenantScope = platformtx.TenantScope

type authorizationQueryer interface {
	QueryRow(context.Context, string, ...any) platformpostgres.Row
}

type boundAuthorizer struct {
	database       authorizationQueryer
	query          string
	includesTenant bool
}

type authorizer struct {
	binder platformpostgres.OwnerBinder[*boundAuthorizer]
	mode   platformtx.ScopeMode
	valid  bool
}

func NewAuthorizer() (authorization.EvidencePort, error) {
	return newAuthorizer("iam", authorizeSQL, false, "")
}

// NewFormsAuthorizer reads IAM-owned evidence through the exact V2 routine while
// remaining inside the caller's Forms transaction capability.
func NewFormsAuthorizer() (authorization.EvidencePort, error) {
	return newAuthorizer("forms", authorizeFormsSQL, true, platformtx.Command)
}

// NewWorkAuthorizer reads current IAM-owned permission evidence while remaining
// inside the caller's Work-owned tenant command transaction.
func NewWorkAuthorizer() (authorization.EvidencePort, error) {
	return newAuthorizer("work", authorizeWorkSQL, true, platformtx.Command)
}

// NewWorkflowHumanTaskAuthorizer reads current IAM evidence through the exact
// Workflow-bound routine inside the caller's actor-bound command scope.
func NewWorkflowHumanTaskAuthorizer() (authorization.EvidencePort, error) {
	return newAuthorizer("workflow", authorizeWorkflowHumanTaskSQL, true, platformtx.Command)
}

// NewFormsSnapshotAuthorizer reads delegated IAM evidence without row locks
// and is sealed to a read-only repeatable-read Forms snapshot transaction.
func NewFormsSnapshotAuthorizer() (authorization.EvidencePort, error) {
	return newAuthorizer(
		"forms", authorizeFormsSnapshotSQL, true, platformtx.SnapshotQuery,
	)
}

// NewPortalAuthorizer restricts evidence to roles owned by the active portal binding.
func NewPortalAuthorizer() (authorization.EvidencePort, error) {
	return newAuthorizer(
		"iam", authorizePortalSQL, true, platformtx.SnapshotQuery,
	)
}

// NewAnalyticsAuthorizer reads current IAM-owned user evidence through the
// exact V2 routine while remaining inside an Analytics-owned tenant scope.
func NewAnalyticsAuthorizer() (authorization.EvidencePort, error) {
	return newAuthorizer("analytics", authorizeAnalyticsSQL, true, "")
}

// NewDocumentsAuthorizer reads current actor evidence inside a Documents-owned
// snapshot used by long-running archive authorization refresh gates.
func NewDocumentsAuthorizer() (authorization.EvidencePort, error) {
	return newAuthorizer("documents", authorizeDocumentsSQL, true,
		platformtx.SnapshotQuery)
}

func newAuthorizer(
	owner string,
	query string,
	includesTenant bool,
	mode platformtx.ScopeMode,
) (authorization.EvidencePort, error) {
	if query == "" {
		return nil, errAuthorizationRead
	}
	binder, err := platformpostgres.NewOwnerBinder(
		owner,
		func(database platformpostgres.BoundExecutor) *boundAuthorizer {
			return &boundAuthorizer{
				database: database, query: query, includesTenant: includesTenant,
			}
		},
	)
	if err != nil {
		return nil, errAuthorizationRead
	}
	return &authorizer{binder: binder, mode: mode, valid: true}, nil
}

var _ authorization.EvidencePort = (*authorizer)(nil)

func (authorizer *authorizer) bind(scope tenantScope) (*boundAuthorizer, error) {
	if authorizer == nil || !authorizer.valid {
		return nil, errAuthorizationRead
	}
	bound, err := platformtx.Bind(scope, authorizer.binder)
	if err != nil || bound == nil {
		return nil, errAuthorizationRead
	}
	return bound, nil
}

func authorizationScopeUsable(
	ctx context.Context,
	scope tenantScope,
	expected platformtx.ScopeMode,
) bool {
	if ctx == nil || ctx.Err() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) {
		return false
	}
	mode := scope.Mode()
	if expected != "" && mode != expected {
		return false
	}
	switch mode {
	case platformtx.Command:
		return platformtx.RequireActive(scope, platformtx.Command) == nil
	case platformtx.SnapshotQuery:
		return platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil
	default:
		return false
	}
}
