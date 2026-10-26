package adoption

import (
	"context"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

func (store *Store) runCommand(ctx context.Context, statement string,
	arguments ...any,
) (application.Run, error) {
	var document []byte
	err := store.within(ctx, func(database Database) error {
		return database.QueryRow(ctx, statement, arguments...).Scan(&document)
	})
	if err != nil {
		return application.Run{}, normalize(err)
	}
	run, err := decodeRun(document)
	return run, normalize(err)
}
