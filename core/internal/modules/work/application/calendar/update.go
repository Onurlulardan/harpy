package calendar

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) UpdateCalendarEvent(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, eventID domain.ID, input CalendarEventInput,
) (ports.OperationReceipt, error) {
	if !eventID.Valid() || meta.ExpectedVersion < 1 || input.Window.Validate() != nil {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	if err := service.checkAudience(ctx, scope, input); err != nil {
		return ports.OperationReceipt{}, err
	}
	hash := ports.Fingerprint(eventID.String(), string(input.Audience), input.BoardID.String(), input.OrganizationID.String(),
		input.Title, input.Description, input.Color, calendarWindowFingerprint(input.Window),
		ports.VersionString(meta.ExpectedVersion))
	return service.mutator.Mutate(ctx, scope, "work.calendar.change", eventID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			event, err := stores.Calendar().LoadEventForUpdate(ctx, eventID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if event.RowVersion != meta.ExpectedVersion || event.Generated || event.DeletedAt != nil {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			event.Audience, event.BoardID, event.OrganizationID = input.Audience, input.BoardID, input.OrganizationID
			event.Title, event.Description, event.Color = input.Title, input.Description, input.Color
			event.Window = input.Window
			event.RowVersion++
			if !event.Valid() {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Calendar().PersistEvent(ctx, event, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				EventID string
				Version int64
			}{eventID.String(), event.RowVersion})
			return ports.OperationReceipt{AggregateID: eventID, RowVersion: event.RowVersion}, payload, nil
		})
}
