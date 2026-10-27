package tasks

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type UpdateInput struct {
	TaskID                       domain.ID
	Priority                     domain.TaskPriority
	Title, Description           string
	StartsAt                     *time.Time
	DueAt                        *time.Time
	Relations                    RelationInput
	ReplaceResource              bool
	Kind                         domain.TaskKind
	Resource                     *ResourceLinkInput
	ReplaceLegacyMetadata        bool
	LegacyMetadata               domain.LegacyTaskMetadata
	CompatibilityImplicitVersion bool
	CompatibilityIntentHash      [32]byte
}

func (service *Service) UpdateTask(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, input UpdateInput) (ports.OperationReceipt, error) {
	input.Title = strings.TrimSpace(input.Title)
	if !input.TaskID.Valid() || !input.Priority.Valid() || input.Title == "" || len(input.Title) > 300 ||
		len(input.Description) > 16_000 || meta.ExpectedVersion < 1 ||
		normalizeRelations(&input.Relations) != nil ||
		!validResourceReplacement(input.ReplaceResource, input.Kind, input.Resource) ||
		(input.ReplaceLegacyMetadata && normalizeLegacyMetadata(&input.LegacyMetadata) != nil) {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	if input.DueAt != nil {
		due := input.DueAt.UTC()
		input.DueAt = &due
	}
	if input.StartsAt != nil {
		starts := input.StartsAt.UTC()
		input.StartsAt = &starts
	}
	assignees, revision := input.Relations.Assignees, int64(0)
	var err error
	if input.Relations.ReplaceAssignees {
		assignees, revision, err = service.checkAssignees(ctx, scope, assignees)
		if err != nil {
			return ports.OperationReceipt{}, err
		}
	}
	var permissionRevisions []int64
	if input.ReplaceResource {
		permissionRevisions, err = service.checkTaskResource(ctx, scope, input.Resource)
		if err != nil {
			return ports.OperationReceipt{}, err
		}
	}
	hash := input.CompatibilityIntentHash
	if input.CompatibilityImplicitVersion {
		if hash == ([32]byte{}) {
			return ports.OperationReceipt{}, ports.ErrInvalid
		}
	} else {
		parts := []string{input.TaskID.String(), input.Title, input.Description,
			string(input.Priority), timeString(input.StartsAt), timeString(input.DueAt),
			ports.VersionString(meta.ExpectedVersion), assigneeFingerprint(assignees, revision),
			resourceReplacementFingerprint(input.ReplaceResource, input.Kind, input.Resource,
				permissionRevisions), legacyMetadataFingerprint(input.ReplaceLegacyMetadata,
				&input.LegacyMetadata)}
		parts = append(parts, relationFingerprint(input.Relations)...)
		hash = ports.Fingerprint(parts...)
	}
	return service.mutator.Mutate(ctx, scope, "work.task.update", input.TaskID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			task, err := stores.Tasks().LoadTaskForUpdate(ctx, input.TaskID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if task.RowVersion != meta.ExpectedVersion || task.DeletedAt != nil {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			if input.ReplaceResource {
				currentLinks, loadErr := stores.Tasks().LoadResourceLinksForUpdate(ctx, task.ID)
				if loadErr != nil {
					return ports.OperationReceipt{}, nil, loadErr
				}
				if !validStoredTaskResource(task, currentLinks) {
					return ports.OperationReceipt{}, nil, ports.ErrUnavailable
				}
				task.Kind = input.Kind
				links, linkErr := service.newResourceLinks(task.ID, input.Resource,
					permissionRevisions)
				if linkErr != nil {
					return ports.OperationReceipt{}, nil, linkErr
				}
				links, linkErr = service.replaceTaskResourceReference(
					ctx, scope, currentLinks, links)
				if linkErr != nil {
					return ports.OperationReceipt{}, nil, linkErr
				}
				if err = stores.Tasks().ReplaceResourceLinks(ctx, task.ID, links); err != nil {
					return ports.OperationReceipt{}, nil, err
				}
			}
			task.Title, task.Description, task.Priority = input.Title, input.Description, input.Priority
			task.StartsAt, task.DueAt = input.StartsAt, input.DueAt
			if err = service.replaceRelations(ctx, stores, task, input.Relations,
				assignees, revision); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			task.RowVersion++
			if !task.Valid() {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Tasks().PersistTask(ctx, task, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if input.ReplaceLegacyMetadata {
				if err = stores.Tasks().UpsertLegacyTaskMetadata(
					ctx, task.ID, input.LegacyMetadata); err != nil {
					return ports.OperationReceipt{}, nil, err
				}
			}
			payload, _ := json.Marshal(struct {
				TaskID  string
				Version int64
			}{task.ID.String(), task.RowVersion})
			return ports.OperationReceipt{AggregateID: task.ID, RowVersion: task.RowVersion}, payload, nil
		})
}
