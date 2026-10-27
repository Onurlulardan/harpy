package tasks

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) ChangeAssignees(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, taskID domain.ID,
	principals []domain.Principal,
) (ports.OperationReceipt, error) {
	if !taskID.Valid() || meta.ExpectedVersion < 1 || len(principals) > 100 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	unique, revision, err := service.checkAssignees(ctx, scope, principals)
	if err != nil {
		return ports.OperationReceipt{}, err
	}
	parts := []string{taskID.String(), ports.VersionString(meta.ExpectedVersion),
		ports.VersionString(revision)}
	for _, principal := range unique {
		parts = append(parts, string(principal.Kind), principal.ID.String())
	}
	return service.mutator.Mutate(ctx, scope, "work.task.assign", taskID, meta,
		ports.Fingerprint(parts...), func(ctx context.Context, stores ports.Stores) (
			ports.OperationReceipt, []byte, error,
		) {
			task, err := stores.Tasks().LoadTaskForUpdate(ctx, taskID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if task.RowVersion != meta.ExpectedVersion || task.DeletedAt != nil {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			allowed, accessErr := stores.Boards().AssigneesHaveAccess(
				ctx, task.BoardID, unique)
			if accessErr != nil || !allowed {
				return ports.OperationReceipt{}, nil, ports.ErrForbidden
			}
			if err = stores.Tasks().ReplaceAssignees(ctx, taskID, unique, revision); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			task.RowVersion++
			if err = stores.Tasks().PersistTask(ctx, task, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				TaskID        string
				AssigneeCount int
				Version       int64
			}{taskID.String(), len(unique), task.RowVersion})
			return ports.OperationReceipt{AggregateID: taskID,
				RowVersion: task.RowVersion}, payload, nil
		})
}

func (service *Service) checkAssignees(ctx context.Context, scope ports.Scope,
	values []domain.Principal,
) ([]domain.Principal, int64, error) {
	seen := map[domain.Principal]struct{}{}
	revision := int64(0)
	for _, principal := range values {
		if !principal.Valid() {
			return nil, 0, ports.ErrInvalid
		}
		if _, exists := seen[principal]; exists {
			return nil, 0, ports.ErrInvalid
		}
		exists, current, err := service.principals.Exists(ctx, scope, principal)
		if err != nil || !exists || current < 1 || revision != 0 && current != revision {
			return nil, 0, ports.ErrForbidden
		}
		revision, seen[principal] = current, struct{}{}
	}
	if len(values) == 0 {
		snapshot, err := service.principals.Resolve(ctx, scope)
		if err != nil || snapshot.PermissionRevision < 1 {
			return nil, 0, ports.ErrForbidden
		}
		revision = snapshot.PermissionRevision
	}
	result := append([]domain.Principal{}, values...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Kind == result[j].Kind {
			return domain.CompareID(result[i].ID, result[j].ID) < 0
		}
		return result[i].Kind < result[j].Kind
	})
	return result, revision, nil
}
