//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type platformLicenseRefreshRequest struct {
	operationID uuid.UUID
	replayed    bool
	createdAt   time.Time
}

type platformLicenseRefreshRead struct {
	operationID      uuid.UUID
	state            string
	snapshotSequence sql.NullInt64
	snapshotSHA256   sql.NullString
	failureCode      sql.NullString
	createdAt        time.Time
	completedAt      sql.NullTime
}

func requestPlatformLicenseRefresh(
	ctx context.Context,
	connection *sql.Conn,
	actorID uuid.UUID,
	key string,
	digest string,
) (platformLicenseRefreshRequest, error) {
	result := platformLicenseRefreshRequest{}
	err := connection.QueryRowContext(ctx, `SELECT * FROM
platform.billing_platform_license_refresh_request_v1($1,$2,$3,$4,$5)`,
		billingEnvironmentID, billingTenantID, actorID, key, digest,
	).Scan(&result.operationID, &result.replayed, &result.createdAt)
	return result, err
}

func readPlatformLicenseRefresh(
	ctx context.Context,
	connection *sql.Conn,
	operationID uuid.UUID,
) (platformLicenseRefreshRead, error) {
	result := platformLicenseRefreshRead{}
	err := connection.QueryRowContext(ctx, `SELECT * FROM
platform.billing_platform_license_refresh_read_v1($1,$2)`,
		billingEnvironmentID, operationID,
	).Scan(
		&result.operationID, &result.state, &result.snapshotSequence,
		&result.snapshotSHA256, &result.failureCode, &result.createdAt,
		&result.completedAt,
	)
	return result, err
}
