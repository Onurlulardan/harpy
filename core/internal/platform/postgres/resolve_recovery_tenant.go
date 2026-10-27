package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/google/uuid"
)

//go:embed adapters/postgres/sql/resolve_recovery_tenant.sql
var resolveRecoveryTenantSQL string

type recoveryTenantRecord struct {
	environmentID uuid.UUID
	tenantID      uuid.UUID
	tenant        TenantRecord
}

func (transaction *RuntimeTransaction) resolveRecoveryTenant(
	ctx context.Context,
	purpose string,
) (recoveryTenantRecord, [32]byte, error) {
	var result recoveryTenantRecord
	var schemaName, slug string
	var digest []byte
	if purpose != "backup" && purpose != "restore" {
		return result, [32]byte{}, ErrRuntimeProfile
	}
	err := transaction.transaction.QueryRow(ctx, resolveRecoveryTenantSQL, purpose).Scan(
		&result.environmentID, &result.tenantID, &slug, &schemaName, &digest,
	)
	if err != nil || len(digest) != 32 {
		return result, [32]byte{}, fmt.Errorf("recovery tenant resolution failed")
	}
	if _, err = tenantSearchPath(schemaName); err != nil {
		return result, [32]byte{}, err
	}
	result.tenant = TenantRecord{schemaName: schemaName, slug: slug}
	var value [32]byte
	copy(value[:], digest)
	return result, value, nil
}
