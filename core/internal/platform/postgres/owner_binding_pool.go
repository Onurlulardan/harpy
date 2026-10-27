package postgres

// PostgresOwnerBindingPool exposes the concrete pool only to module-owned
// PostgreSQL bind adapters. Business module roots depend on adapter contracts.
func (pool *Pool) PostgresOwnerBindingPool() *Pool {
	return pool
}
