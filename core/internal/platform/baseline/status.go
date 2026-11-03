package baseline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func Inspect(ctx context.Context, dsn string) (bool, error) {
	config, err := sessionConfig(dsn, "", true)
	if err != nil {
		return false, err
	}
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return false, fmt.Errorf("open baseline database: %w", err)
	}
	defer conn.Close(ctx)
	var applied bool
	err = conn.QueryRow(ctx, `SELECT EXISTS (
SELECT 1 FROM platform.goose_db_version WHERE version_id = $1 AND is_applied
)`, platformVersion).Scan(&applied)
	if err != nil {
		return false, nil
	}
	return applied, nil
}

func PlatformDigest() string {
	sum := sha256.Sum256([]byte(platformScript))
	return hex.EncodeToString(sum[:])
}
