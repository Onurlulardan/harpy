package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewContractRetirementRuntime is the API composition store factory.
func NewContractRetirementRuntime(
	pool *pgxpool.Pool,
	_ time.Duration,
) (*ContractRetirementDatabase, error) {
	return NewContractRetirementDatabase(pool)
}
