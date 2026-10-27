package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type roleCatalog struct {
	name        string
	superuser   bool
	inherit     bool
	createRole  bool
	createDB    bool
	login       bool
	replication bool
	bypassRLS   bool
	connLimit   int
}

func validateRolePreflight(
	ctx context.Context,
	queryer roleQueryer,
	contract profileContract,
) error {
	var sessionRole, currentRole roleCatalog
	var admin, inherit, set bool
	var membershipCount int
	var sessionTemp, currentTemp, currentCreate bool
	var sessionConnect, currentConnect bool
	err := queryer.QueryRow(ctx, preflightRolesSQL).Scan(
		&sessionRole.name, &sessionRole.superuser, &sessionRole.inherit,
		&sessionRole.createRole, &sessionRole.createDB, &sessionRole.login,
		&sessionRole.replication, &sessionRole.bypassRLS, &sessionRole.connLimit,
		&currentRole.name, &currentRole.superuser, &currentRole.inherit,
		&currentRole.createRole, &currentRole.createDB, &currentRole.login,
		&currentRole.replication, &currentRole.bypassRLS, &currentRole.connLimit,
		&admin, &inherit, &set, &membershipCount,
		&sessionTemp, &currentTemp, &sessionConnect, &currentConnect, &currentCreate,
	)
	if err != nil {
		return fmt.Errorf("%w: role query", ErrStartupPreflight)
	}
	if !safeRole(sessionRole, contract.sessionUser, true) ||
		!safeRole(currentRole, contract.currentUser, contract.currentLogin) ||
		admin || inherit || set != contract.memberSet ||
		membershipCount != contract.memberships ||
		sessionTemp || (!contract.currentTemp && currentTemp) ||
		!sessionConnect || !currentConnect || currentCreate != contract.currentCreate {
		return fmt.Errorf("%w: role contract", ErrStartupPreflight)
	}
	return nil
}

type roleQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func safeRole(role roleCatalog, name string, login bool) bool {
	return role.name == name && role.login == login && !role.superuser &&
		!role.inherit && !role.createRole && !role.createDB &&
		!role.replication && !role.bypassRLS && role.connLimit == -1
}
