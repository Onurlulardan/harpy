package postgres

import "errors"

var (
	ErrPoolClosed            = errors.New("postgres pool is closed")
	ErrSessionTainted        = errors.New("postgres session is tainted")
	ErrStartupPreflight      = errors.New("postgres startup preflight failed")
	ErrRuntimeProfile        = errors.New("postgres pool is not a runtime profile")
	ErrSchemaExecutorProfile = errors.New("postgres pool is not the schema executor profile")
	ErrTransactionClosed     = errors.New("postgres runtime transaction is closed")
	ErrCommitRolledBack      = errors.New("postgres transaction commit was rolled back")
	ErrCommitAmbiguous       = errors.New("postgres transaction commit outcome is ambiguous")
	ErrBindingAbandoned      = errors.New("postgres bound operation escaped its runtime scope")
	ErrBindingCleanupTimeout = errors.New("postgres bound operation cleanup timed out")
)
