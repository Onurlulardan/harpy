package recurrence

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type RecurrenceInput struct {
	EventID            domain.ID
	DTSTART, TimeZone  string
	TZDBVersion, RRULE string
	EXDATE             []string
}

func (service *Service) CreateRecurrence(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, input RecurrenceInput,
) (ports.OperationReceipt, error) {
	if !input.EventID.Valid() || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	definition, revision, err := service.parseRecurrence(ctx, input, 1)
	if err != nil {
		return ports.OperationReceipt{}, err
	}
	hash := recurrenceFingerprint(input, meta.ExpectedVersion, 1)
	return service.mutator.Mutate(ctx, scope, "work.recurrence.change", input.EventID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			event, err := stores.Calendar().LoadEventForUpdate(ctx, input.EventID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if event.RowVersion != meta.ExpectedVersion || event.Generated || event.DeletedAt != nil {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			ruleID, err := service.ids.New(service.clock.Now().UTC())
			if err != nil {
				return ports.OperationReceipt{}, nil, ports.ErrUnavailable
			}
			if err = stores.Calendar().InsertRecurrenceRule(ctx, ruleID, input.EventID, revision); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			event.RowVersion++
			if err = stores.Calendar().PersistEvent(ctx, event, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				EventID, RuleID, Contract string
				Version                   int64
			}{
				input.EventID.String(), ruleID.String(), definition.Contract, event.RowVersion})
			return ports.OperationReceipt{AggregateID: input.EventID, RowVersion: event.RowVersion}, payload, nil
		})
}

func (service *Service) parseRecurrence(ctx context.Context, input RecurrenceInput, revision int64) (domain.RecurrenceDefinition, ports.RecurrenceRevision, error) {
	location, err := service.tzdb.Load(ctx, input.TimeZone, input.TZDBVersion)
	if err != nil {
		return domain.RecurrenceDefinition{}, ports.RecurrenceRevision{}, ports.ErrInvalid
	}
	definition, err := domain.ParseRecurrence(domain.RecurrenceInput{DTStart: input.DTSTART,
		TimeZone: input.TimeZone, TZDBVersion: input.TZDBVersion, RRULE: input.RRULE,
		EXDATE: append([]string{}, input.EXDATE...), Location: location})
	if err != nil {
		return domain.RecurrenceDefinition{}, ports.RecurrenceRevision{}, err
	}
	record := ports.RecurrenceRevision{Revision: revision, Contract: domain.RecurrenceContract,
		DTSTART: input.DTSTART, TimeZone: input.TimeZone, TZDBVersion: input.TZDBVersion,
		RRULE: input.RRULE, EXDATE: append([]string{}, input.EXDATE...)}
	return definition, record, nil
}

func recurrenceFingerprint(input RecurrenceInput, eventVersion, revision int64) [32]byte {
	parts := []string{input.EventID.String(), input.DTSTART, input.TimeZone, input.TZDBVersion,
		input.RRULE, ports.VersionString(eventVersion), ports.VersionString(revision)}
	parts = append(parts, input.EXDATE...)
	return ports.Fingerprint(parts...)
}
