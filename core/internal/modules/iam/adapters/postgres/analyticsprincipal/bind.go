package analyticsprincipal

import (
	"context"
	"errors"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var ErrPermissionRead = errors.New("IAM Analytics principal permission read failed")

type queryer interface {
	QueryRow(context.Context, string, ...any) platformpostgres.Row
}

type boundPort struct{ database queryer }

type Reader interface {
	RoleAllowed(context.Context, platformtx.TenantScope,
		uuid.UUID, string, string) (bool, error)
}

type roleReader func(context.Context, platformtx.TenantScope,
	uuid.UUID, string, string) (bool, error)

type reader struct{ roleAllowed roleReader }

type port struct {
	binder platformpostgres.OwnerBinder[*boundPort]
	valid  bool
}

func New() (Reader, error) {
	binder, err := platformpostgres.NewOwnerBinder("analytics",
		func(database platformpostgres.BoundExecutor) *boundPort {
			return &boundPort{database: database}
		})
	if err != nil || rolePermissionSQL == "" {
		return nil, ErrPermissionRead
	}
	bound := &port{binder: binder, valid: true}
	return &reader{roleAllowed: bound.roleAllowed}, nil
}

func (port *port) roleAllowed(
	ctx context.Context, scope platformtx.TenantScope,
	roleID uuid.UUID, resource, action string,
) (bool, error) {
	if port == nil || !port.valid || ctx == nil || ctx.Err() != nil ||
		roleID == uuid.Nil || resource == "" || action == "" ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireOwner(scope, "analytics") != nil {
		return false, ErrPermissionRead
	}
	bound, err := platformtx.Bind(scope, port.binder)
	if err != nil || bound == nil || bound.database == nil {
		return false, ErrPermissionRead
	}
	var allowed bool
	err = bound.database.QueryRow(ctx, rolePermissionSQL, scope.TenantID(),
		roleID, resource, action).Scan(&allowed)
	if err != nil {
		return false, ErrPermissionRead
	}
	return allowed, nil
}

func (value *reader) RoleAllowed(ctx context.Context, scope platformtx.TenantScope,
	roleID uuid.UUID, resource, action string) (bool, error) {
	if value == nil || value.roleAllowed == nil {
		return false, ErrPermissionRead
	}
	return value.roleAllowed(ctx, scope, roleID, resource, action)
}

var _ Reader = (*reader)(nil)
