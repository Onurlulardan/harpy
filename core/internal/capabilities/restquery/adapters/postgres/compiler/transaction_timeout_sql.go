package compiler

const transactionTimeoutSQL = "SELECT pg_catalog.set_config('statement_timeout', $1, true)"

// TransactionTimeoutSQL is the fixed transaction-local timeout statement.
func TransactionTimeoutSQL() string { return transactionTimeoutSQL }
