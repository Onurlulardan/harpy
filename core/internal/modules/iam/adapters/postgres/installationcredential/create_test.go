package installationcredentialpostgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	installationcredential "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/installationcredential"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestCreateUsesOwnerRoutineWithoutPlaintext(t *testing.T) {
	if !strings.Contains(createSQL, "iam_installation_credential_initialize_v2") ||
		strings.Contains(strings.ToLower(createSQL), "password") {
		t.Fatal("installation credential initialize SQL contract drifted")
	}
}

func TestCreateReportsSQLState(t *testing.T) {
	store := &store{queryRow: func(context.Context, string, ...any) Row {
		return failRow{err: &pgconn.PgError{
			Code: "42501", Message: "permission denied for schema platform",
			Detail: "activation routine",
		}}
	}}
	_, err := store.Create(context.Background(), installationcredential.ActivationBinding{},
		installationcredential.CreateCommand{})
	if !errors.Is(err, installationcredential.ErrStorage) ||
		!strings.Contains(err.Error(), "SQLSTATE 42501") ||
		!strings.Contains(err.Error(), "permission denied for schema platform") ||
		!strings.Contains(err.Error(), "activation routine") {
		t.Fatal(err)
	}
}

type failRow struct{ err error }

func (row failRow) Scan(...any) error { return row.err }
