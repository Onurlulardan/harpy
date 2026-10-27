package adoption

import (
	"context"
	"encoding/json"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

func (store *Store) AdoptionExecutePage(ctx context.Context, command application.ApplyPageCommand) (
	application.PageReceipt, error,
) {
	rows, err := json.Marshal(command.Records)
	if err != nil {
		return application.PageReceipt{}, application.ErrInvalid
	}
	statement := applyContactSQL
	if command.Kind == application.BugTenant {
		statement = applyBugSQL
	} else if command.Kind != application.ContactGlobal {
		return application.PageReceipt{}, application.ErrInvalid
	}
	var document []byte
	err = store.within(ctx, func(database Database) error {
		return database.QueryRow(ctx, statement, command.OperationID, command.RunID,
			command.LeaseOwner, command.LeaseGeneration, command.PageNumber,
			nullableUUID(command.CursorBefore), rows, command.SourcePageSHA256[:],
			command.ObservedAt).Scan(&document)
	})
	if err != nil {
		return application.PageReceipt{}, normalize(err)
	}
	receipt, err := decodePage(document)
	return receipt, normalize(err)
}
