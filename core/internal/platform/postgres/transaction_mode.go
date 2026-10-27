package postgres

import "github.com/jackc/pgx/v5"

type TransactionMode string

const (
	TransactionCommand           TransactionMode = "command"
	TransactionRepeatableCommand TransactionMode = "repeatable_command"
	TransactionSnapshotQuery     TransactionMode = "snapshot_query"
)

func (mode TransactionMode) options() (pgx.TxOptions, bool) {
	switch mode {
	case TransactionCommand:
		return pgx.TxOptions{
			IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite,
			DeferrableMode: pgx.NotDeferrable,
		}, true
	case TransactionRepeatableCommand:
		return pgx.TxOptions{
			IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadWrite,
			DeferrableMode: pgx.NotDeferrable,
		}, true
	case TransactionSnapshotQuery:
		return pgx.TxOptions{
			IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly,
			DeferrableMode: pgx.NotDeferrable,
		}, true
	default:
		return pgx.TxOptions{}, false
	}
}
