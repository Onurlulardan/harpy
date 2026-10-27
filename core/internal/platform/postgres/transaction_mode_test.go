package postgres

import (
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestRepeatableCommandIsReadWriteRepeatableRead(t *testing.T) {
	options, valid := TransactionRepeatableCommand.options()
	if !valid || options.IsoLevel != pgx.RepeatableRead ||
		options.AccessMode != pgx.ReadWrite ||
		options.DeferrableMode != pgx.NotDeferrable {
		t.Fatalf("repeatable command options = %+v, valid=%v", options, valid)
	}
}
