package postgres

import (
	"context"
	"errors"
	"regexp"
	"slices"

	"github.com/google/uuid"
)

var snapshotDomainPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

// ResolveSnapshotDomainGenerations reads a closed platform projection. The
// resulting generations are revalidated inside the caller's TenantScope.
func (pool *Pool) ResolveSnapshotDomainGenerations(
	ctx context.Context,
	tenantID uuid.UUID,
	domains []string,
) (map[string]int64, error) {
	if pool == nil || ctx == nil || tenantID == uuid.Nil ||
		!validSnapshotDomains(domains) {
		return nil, ErrAuthorityMismatch
	}
	database, err := pool.BeginRuntimeTransaction(ctx, TransactionSnapshotQuery)
	if err != nil {
		return nil, err
	}
	rows, err := database.transaction.Query(
		ctx, resolveSnapshotDomainsSQL, tenantID, domains[0], domains[1],
	)
	if err != nil {
		_ = database.Rollback()
		return nil, err
	}
	result := make(map[string]int64, len(domains))
	for rows.Next() {
		var domain string
		var generation int64
		if err := rows.Scan(&domain, &generation); err != nil ||
			!snapshotDomainPattern.MatchString(domain) || generation <= 0 {
			rows.Close()
			_ = database.Rollback()
			return nil, ErrAuthorityMismatch
		}
		result[domain] = generation
	}
	err = rows.Err()
	rows.Close()
	if err != nil || len(result) != len(domains) {
		_ = database.Rollback()
		return nil, errors.Join(ErrAuthorityMismatch, err)
	}
	if err := database.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func validSnapshotDomains(domains []string) bool {
	if len(domains) != 2 {
		return false
	}
	copyOfDomains := append([]string(nil), domains...)
	slices.Sort(copyOfDomains)
	for index, domain := range copyOfDomains {
		if !snapshotDomainPattern.MatchString(domain) ||
			index > 0 && copyOfDomains[index-1] == domain {
			return false
		}
	}
	return true
}
