package calendar

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type CalendarEventInput struct {
	Audience                domain.Audience
	BoardID, OrganizationID domain.ID
	Title, Description      string
	Color                   string
	Window                  domain.CalendarWindow
}

func (service *Service) CreateCalendarEvent(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, input CalendarEventInput,
) (ports.OperationReceipt, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" || len(input.Title) > 300 || len(input.Description) > 16_000 ||
		!domain.ValidColor(input.Color) || meta.ExpectedVersion != 0 || input.Window.Validate() != nil {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	if err := service.checkAudience(ctx, scope, input); err != nil {
		return ports.OperationReceipt{}, err
	}
	resource := scope.ActorID()
	if input.Audience == domain.AudienceBoard {
		resource = input.BoardID
	} else if input.Audience == domain.AudienceOrganization {
		resource = input.OrganizationID
	}
	hash := ports.Fingerprint(string(input.Audience), input.BoardID.String(), input.OrganizationID.String(),
		input.Title, input.Description, input.Color, calendarWindowFingerprint(input.Window))
	return service.mutator.Mutate(ctx, scope, "work.calendar.change", resource, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			if input.Audience == domain.AudienceBoard {
				board, err := stores.Boards().LoadBoardForUpdate(ctx, input.BoardID)
				if err != nil {
					return ports.OperationReceipt{}, nil, err
				}
				if board.Lifecycle != domain.LifecycleActive {
					return ports.OperationReceipt{}, nil, ports.ErrConflict
				}
			}
			id, err := service.ids.New(service.clock.Now().UTC())
			if err != nil {
				return ports.OperationReceipt{}, nil, ports.ErrUnavailable
			}
			event := domain.CalendarEvent{ID: id, OwnerID: scope.ActorID(),
				BoardID: input.BoardID, OrganizationID: input.OrganizationID,
				Title: input.Title, Description: input.Description, Color: input.Color,
				Audience: input.Audience,
				Window:   input.Window, RowVersion: 1}
			if !event.Valid() {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Calendar().InsertEvent(ctx, event); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct{ EventID, Audience string }{id.String(), string(input.Audience)})
			return ports.OperationReceipt{AggregateID: id, RowVersion: 1}, payload, nil
		})
}

func (service *Service) checkAudience(ctx context.Context, scope ports.Scope, input CalendarEventInput) error {
	switch input.Audience {
	case domain.AudiencePrivate:
		if input.BoardID.Valid() || input.OrganizationID.Valid() {
			return ports.ErrInvalid
		}
	case domain.AudienceBoard:
		if !input.BoardID.Valid() || input.OrganizationID.Valid() {
			return ports.ErrInvalid
		}
	case domain.AudienceOrganization:
		if input.BoardID.Valid() || !input.OrganizationID.Valid() {
			return ports.ErrInvalid
		}
		exists, revision, err := service.principals.Exists(ctx, scope,
			domain.Principal{Kind: domain.PrincipalOrganization, ID: input.OrganizationID})
		if err != nil || !exists || revision < 1 {
			return ports.ErrForbidden
		}
	default:
		return ports.ErrInvalid
	}
	return nil
}

func timeString(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func calendarWindowFingerprint(window domain.CalendarWindow) string {
	return timeString(window.StartsAt) + ":" + timeString(window.EndsAt) + ":" +
		window.StartDate + ":" + window.EndDate + ":" + window.TimeZone
}
