package v1cutover

import app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"

type scanner interface{ Scan(...any) error }

func scanAuthority(row scanner) (app.Authority, error) {
	var value app.Authority
	err := row.Scan(&value.Writer, &value.State, &value.Generation, &value.RowVersion)
	if err != nil || !value.Valid() {
		return app.Authority{}, app.ErrConflict
	}
	return value, nil
}
