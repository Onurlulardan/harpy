package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func (store *Store) ContactSubmit(ctx context.Context,
	command application.ContactSubmission,
) (domain.ContactMessage, bool, error) {
	var message domain.ContactMessage
	var replayed, admitted bool
	err := store.within(ctx, CommandTransaction, func(database Database) error {
		var document []byte
		err := database.QueryRow(ctx, submitContactSQL, uuid.UUID(command.ID), command.Input.Name,
			command.Input.Email, command.Input.Company, command.Input.Phone,
			command.Input.EmployeeCount, command.Input.Message, command.Input.ContactType,
			command.Input.Source, command.RateKeySHA256[:], command.IdempotencyKey,
			command.RequestSHA256[:], command.ObservedAt).Scan(&document, &replayed, &admitted)
		if err != nil || !admitted {
			return err
		}
		message, err = decodeContact(document)
		return err
	})
	if err != nil {
		return domain.ContactMessage{}, false, store.normalize(err)
	}
	if !admitted {
		return domain.ContactMessage{}, false, application.ErrRateLimited
	}
	return message, replayed, nil
}
