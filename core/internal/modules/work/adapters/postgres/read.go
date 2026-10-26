package postgres

import (
	"context"
	_ "embed"
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/tasks"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

//go:embed sql/tenant/read_visible_records.sql
var readVisibleRecordsSQL string

//go:embed sql/tenant/read_task_resource_links.sql
var readTaskResourceLinksSQL string

type Reader struct {
	principals ports.PrincipalResolver
	access     *ownerAccess
}

type ReadOptions struct {
	Status               domain.TaskStatus
	Priority             domain.TaskPriority
	BoardID              domain.ID
	StartsAt, EndsAt     *time.Time
	OwnerOnly, Ascending bool
	Compatibility        bool
}

func NewReader(principals ports.PrincipalResolver) (*Reader, error) {
	if principals == nil || readVisibleRecordsSQL == "" || readTaskResourceLinksSQL == "" {
		return nil, application.ErrUnavailable
	}
	access, err := newOwnerAccess()
	if err != nil {
		return nil, application.ErrUnavailable
	}
	return &Reader{principals: principals, access: access}, nil
}

func (reader *Reader) Read(ctx context.Context, local ports.Scope, resource string,
	targetID, parentID domain.ID, limit int, options ReadOptions,
) (json.RawMessage, error) {
	bridge, ok := local.(*scopeBridge)
	if reader == nil || !ok || !scopeUsable(ctx, bridge) || limit < 1 || limit > 500 {
		return nil, application.ErrInvalid
	}
	snapshot, err := reader.principals.Resolve(ctx, local)
	if err != nil {
		return nil, application.ErrForbidden
	}
	kinds, ids := principalArrays(snapshot.Principals)
	database, err := reader.access.bind(ctx, bridge)
	if err != nil {
		return nil, application.ErrUnavailable
	}
	var document []byte
	err = database.QueryRow(ctx, readVisibleRecordsSQL, resource,
		bridge.scope.ActorID(), kinds, ids, uuid.UUID(targetID),
		uuid.UUID(parentID), limit, string(options.Status), uuid.UUID(options.BoardID),
		options.StartsAt, options.EndsAt, options.OwnerOnly, options.Ascending,
		string(options.Priority), options.Compatibility).Scan(&document)
	if err != nil || !json.Valid(document) {
		return nil, application.ErrUnavailable
	}
	return append(json.RawMessage(nil), document...), nil
}

func (reader *Reader) ReadTaskResources(
	ctx context.Context, local ports.Scope, taskIDs []domain.ID,
) (map[domain.ID]*tasks.ResourceLinkInput, error) {
	bridge, ok := local.(*scopeBridge)
	if reader == nil || !ok || !scopeUsable(ctx, bridge) ||
		len(taskIDs) > domain.MaxBoardTasks {
		return nil, application.ErrInvalid
	}
	requested := make(map[domain.ID]struct{}, len(taskIDs))
	values := make([]uuid.UUID, len(taskIDs))
	for index, taskID := range taskIDs {
		if !taskID.Valid() {
			return nil, application.ErrInvalid
		}
		if _, duplicate := requested[taskID]; duplicate {
			return nil, application.ErrInvalid
		}
		requested[taskID], values[index] = struct{}{}, uuid.UUID(taskID)
	}
	result := make(map[domain.ID]*tasks.ResourceLinkInput, len(taskIDs))
	if len(values) == 0 {
		return result, nil
	}
	database, err := reader.access.bind(ctx, bridge)
	if err != nil {
		return nil, application.ErrUnavailable
	}
	queryRows, err := database.Query(ctx, readTaskResourceLinksSQL, values)
	if err != nil || queryRows == nil {
		return nil, application.ErrUnavailable
	}
	defer queryRows.Close()
	links, err := scanTaskResourceLinks(queryRows, requested)
	if err != nil {
		return nil, err
	}
	for _, taskID := range taskIDs {
		resource, resourceErr := tasks.ResourceInputFromLinks(taskID, links[taskID])
		if resourceErr != nil {
			return nil, application.ErrUnavailable
		}
		result[taskID] = resource
	}
	return result, nil
}

func scanTaskResourceLinks(
	queryRows rows, requested map[domain.ID]struct{},
) (map[domain.ID][]domain.ResourceLink, error) {
	result := make(map[domain.ID][]domain.ResourceLink, len(requested))
	expected := make(map[domain.ID]int64, len(requested))
	for queryRows.Next() {
		var id, workID, targetID, versionID uuid.UUID
		var owner, kind string
		var epoch, permissionRevision, rowVersion, linkCount int64
		if err := queryRows.Scan(&id, &workID, &owner, &kind, &targetID,
			&versionID, &epoch, &permissionRevision, &rowVersion, &linkCount); err != nil {
			return nil, application.ErrUnavailable
		}
		link := domain.ResourceLink{ID: domain.ID(id), WorkResourceID: domain.ID(workID),
			TargetOwner: domain.SourceOwner(owner), TargetKind: kind,
			TargetID: domain.ID(targetID), TargetVersionID: domain.ID(versionID),
			ReferenceEpoch: epoch, PermissionRevision: permissionRevision,
			RowVersion: rowVersion}
		if _, ok := requested[link.WorkResourceID]; !ok || !link.Valid() ||
			linkCount < 1 || linkCount > 2 || len(result[link.WorkResourceID]) >= 2 {
			return nil, application.ErrUnavailable
		}
		if count, seen := expected[link.WorkResourceID]; seen && count != linkCount {
			return nil, application.ErrUnavailable
		}
		expected[link.WorkResourceID] = linkCount
		result[link.WorkResourceID] = append(result[link.WorkResourceID], link)
	}
	if queryRows.Err() != nil {
		return nil, application.ErrUnavailable
	}
	for taskID, count := range expected {
		if int64(len(result[taskID])) != count {
			return nil, application.ErrUnavailable
		}
	}
	return result, nil
}
