package postgres

import (
	"context"
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

type sessionObservation struct {
	pid              int64
	backendStart     time.Time
	sessionUser      string
	applicationName  string
	systemIdentifier string
	databaseOID      uint32
	databaseName     string
}

func observeSession(ctx context.Context, connection *pgx.Conn) (sessionObservation, error) {
	var value sessionObservation
	var databaseOID int64
	err := connection.QueryRow(ctx, observeSessionSQL).Scan(
		&value.pid, &value.backendStart, &value.sessionUser, &value.applicationName,
		&value.systemIdentifier, &databaseOID, &value.databaseName,
	)
	if err != nil || databaseOID < 1 || databaseOID > 4294967295 {
		return sessionObservation{}, fmt.Errorf("observe cutover PostgreSQL session")
	}
	value.databaseOID = uint32(databaseOID)
	return value, nil
}

func (value sessionObservation) validateDatabase(expected cutoverv1.DatabaseIdentity) error {
	if value.systemIdentifier != expected.SystemIdentifier ||
		value.databaseOID != expected.DatabaseOID || value.databaseName != expected.DatabaseName {
		return fmt.Errorf("cutover PostgreSQL database identity differs")
	}
	return nil
}

func (value sessionObservation) matches(expected cutoverv1.CutoverSession) bool {
	backendStart, err := time.Parse(time.RFC3339Nano, expected.BackendStart)
	return err == nil && value.pid == expected.PID && value.backendStart.Equal(backendStart) &&
		value.sessionUser == expected.SessionUser && value.applicationName == expected.ApplicationName
}
