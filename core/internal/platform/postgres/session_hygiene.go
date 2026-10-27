package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func validateSession(
	ctx context.Context,
	connection *pgx.Conn,
	contract profileContract,
) error {
	if connection.PgConn().TxStatus() != 'I' {
		return ErrSessionTainted
	}
	return validateSessionState(ctx, connection, contract)
}

func validateSessionState(
	ctx context.Context,
	queryer roleQueryer,
	contract profileContract,
) error {
	var sessionUser, currentUser, applicationName string
	var timezone, searchPath, encoding, replicationRole, rowSecurity string
	var tempSchemaOID int64
	var primary, noAdvisoryLocks, noListenChannels bool
	err := queryer.QueryRow(ctx, sessionHygieneSQL).Scan(
		&sessionUser, &currentUser, &applicationName, &timezone, &searchPath,
		&encoding, &replicationRole, &rowSecurity, &tempSchemaOID, &primary,
		&noAdvisoryLocks, &noListenChannels,
	)
	if err != nil {
		return err
	}
	if sessionUser != contract.sessionUser || currentUser != contract.currentUser ||
		applicationName != contract.applicationName || timezone != "UTC" ||
		searchPath != "pg_catalog" || encoding != "UTF8" ||
		replicationRole != "origin" || rowSecurity != "on" || tempSchemaOID != 0 ||
		!primary || !noAdvisoryLocks || !noListenChannels {
		return ErrSessionTainted
	}
	return nil
}

func reusableSession(
	ctx context.Context,
	connection *pgx.Conn,
	contract profileContract,
) (bool, error) {
	err := validateSession(ctx, connection, contract)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrSessionTainted) {
		return false, nil
	}
	return false, err
}
