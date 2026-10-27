package tasks

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type ResourceTargetInput struct {
	Kind          string
	ID, VersionID domain.ID
}

type ResourceLinkInput struct {
	Owner   domain.SourceOwner
	Targets []ResourceTargetInput
}

type CreateTaskInput struct {
	BoardID, ColumnID  domain.ID
	Kind               domain.TaskKind
	Priority           domain.TaskPriority
	Title, Description string
	StartsAt           *time.Time
	DueAt              *time.Time
	Before, After      *domain.ID
	Assignees          []domain.Principal
	Labels             []string
	Checklist          []ChecklistValue
	Resource           *ResourceLinkInput
	LegacyMetadata     *domain.LegacyTaskMetadata
}

func (service *Service) CreateTask(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, input CreateTaskInput,
) (ports.OperationReceipt, error) {
	input.Title = strings.TrimSpace(input.Title)
	relations := RelationInput{Assignees: input.Assignees, Labels: input.Labels,
		Checklist: input.Checklist}
	if !input.BoardID.Valid() || !input.ColumnID.Valid() || !input.Kind.Valid() ||
		!input.Priority.Valid() ||
		input.Title == "" || len(input.Title) > 300 || len(input.Description) > 16_000 ||
		meta.ExpectedVersion != 0 || !resourceMatchesKind(input.Kind, input.Resource) ||
		normalizeRelations(&relations) != nil ||
		normalizeLegacyMetadata(input.LegacyMetadata) != nil {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	input.Assignees, input.Labels, input.Checklist = relations.Assignees,
		relations.Labels, relations.Checklist
	if input.DueAt != nil {
		due := input.DueAt.UTC()
		input.DueAt = &due
	}
	if input.StartsAt != nil {
		starts := input.StartsAt.UTC()
		input.StartsAt = &starts
	}
	permissionRevisions, err := service.checkTaskResource(ctx, scope, input.Resource)
	if err != nil {
		return ports.OperationReceipt{}, err
	}
	assignees, assigneeRevision, err := service.checkAssignees(ctx, scope, input.Assignees)
	if err != nil {
		return ports.OperationReceipt{}, err
	}
	parts := []string{input.BoardID.String(), input.ColumnID.String(), string(input.Kind),
		string(input.Priority), input.Title, input.Description, timeString(input.StartsAt),
		timeString(input.DueAt), resourceAuthorityFingerprint(input.Resource,
			permissionRevisions),
		assigneeFingerprint(assignees, assigneeRevision)}
	parts = append(parts, legacyMetadataFingerprint(input.LegacyMetadata != nil,
		input.LegacyMetadata))
	parts = append(parts, relationFingerprint(relations)...)
	hash := ports.Fingerprint(parts...)
	return service.mutator.Mutate(ctx, scope, "work.task.create", input.BoardID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			board, err := stores.Boards().LoadBoardForUpdate(ctx, input.BoardID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if board.Lifecycle != domain.LifecycleActive {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			column, err := stores.Boards().LoadColumnForUpdate(ctx, input.ColumnID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if column.BoardID != input.BoardID {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			allowed, accessErr := stores.Boards().AssigneesHaveAccess(
				ctx, input.BoardID, assignees)
			if accessErr != nil || !allowed {
				return ports.OperationReceipt{}, nil, ports.ErrForbidden
			}
			count, err := stores.Tasks().CountBoardTasksForUpdate(ctx, input.BoardID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if count >= domain.MaxBoardTasks {
				return ports.OperationReceipt{}, nil, domain.ErrBoardLimit
			}
			ranked, err := stores.Tasks().LoadRankedTasksForUpdate(ctx, input.ColumnID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			id, err := service.ids.New(service.clock.Now().UTC())
			if err != nil {
				return ports.OperationReceipt{}, nil, ports.ErrUnavailable
			}
			plan, err := domain.PlanPosition(ranked, id, domain.MoveNeighbors{Before: input.Before, After: input.After})
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if len(plan.Rebalanced) > 0 {
				if err = stores.Tasks().RebalanceTasks(ctx, input.ColumnID, plan.Rebalanced); err != nil {
					return ports.OperationReceipt{}, nil, err
				}
			}
			task := domain.Task{ID: id, BoardID: input.BoardID, ColumnID: input.ColumnID,
				CreatorID: scope.ActorID(),
				Kind:      input.Kind, Status: column.Status, Priority: input.Priority, Title: input.Title,
				Description: input.Description, Position: plan.Position, RowVersion: 1, DueAt: input.DueAt}
			task.StartsAt = input.StartsAt
			if !task.Valid() {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Tasks().InsertTask(ctx, task); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if input.LegacyMetadata != nil {
				if err = stores.Tasks().UpsertLegacyTaskMetadata(
					ctx, id, *input.LegacyMetadata); err != nil {
					return ports.OperationReceipt{}, nil, err
				}
			}
			if len(assignees) > 0 {
				if err = stores.Tasks().ReplaceAssignees(ctx, id,
					assignees, assigneeRevision); err != nil {
					return ports.OperationReceipt{}, nil, err
				}
			}
			if len(input.Labels) > 0 {
				if err = service.replaceLabels(ctx, stores.Tasks(), task, input.Labels); err != nil {
					return ports.OperationReceipt{}, nil, err
				}
			}
			if len(input.Checklist) > 0 {
				if err = service.replaceChecklist(ctx, stores.Tasks(), id,
					input.Checklist); err != nil {
					return ports.OperationReceipt{}, nil, err
				}
			}
			if input.Resource != nil {
				links, linkErr := service.newResourceLinks(
					id, input.Resource, permissionRevisions)
				if linkErr != nil {
					return ports.OperationReceipt{}, nil, linkErr
				}
				links, linkErr = service.registerTaskResource(ctx, scope, links)
				if linkErr != nil {
					return ports.OperationReceipt{}, nil, linkErr
				}
				for _, link := range links {
					if err = stores.Tasks().InsertResourceLink(ctx, link); err != nil {
						return ports.OperationReceipt{}, nil, err
					}
				}
			}
			payload, _ := json.Marshal(struct {
				BoardID, TaskID string
				Version         int64
			}{
				input.BoardID.String(), id.String(), 1})
			return ports.OperationReceipt{AggregateID: id, RowVersion: 1}, payload, nil
		})
}
