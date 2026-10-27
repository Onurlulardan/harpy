package recurrence

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) UpdateRecurrence(
	ctx context.Context, scope ports.Scope, meta ports.Mutation,
	input RecurrenceInput, expectedRuleRevision int64,
) (ports.OperationReceipt, error) {
	if !input.EventID.Valid() || meta.ExpectedVersion < 1 || expectedRuleRevision < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	_, revision, err := service.parseRecurrence(ctx, input, expectedRuleRevision+1)
	if err != nil {
		return ports.OperationReceipt{}, err
	}
	hash := recurrenceFingerprint(input, meta.ExpectedVersion, expectedRuleRevision)
	return service.mutator.Mutate(ctx, scope, "work.recurrence.change", input.EventID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			event, err := stores.Calendar().LoadEventForUpdate(ctx, input.EventID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if event.RowVersion != meta.ExpectedVersion || event.Generated || event.DeletedAt != nil {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			ruleID, current, err := stores.Calendar().FindRecurrenceRuleForEvent(ctx, input.EventID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if current != expectedRuleRevision {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			if err = stores.Calendar().AppendRecurrenceRevision(ctx, ruleID, revision, expectedRuleRevision); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			event.RowVersion++
			if err = stores.Calendar().PersistEvent(ctx, event, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				EventID, RuleID       string
				RuleRevision, Version int64
			}{
				input.EventID.String(), ruleID.String(), revision.Revision, event.RowVersion})
			return ports.OperationReceipt{AggregateID: input.EventID, RowVersion: event.RowVersion}, payload, nil
		})
}
