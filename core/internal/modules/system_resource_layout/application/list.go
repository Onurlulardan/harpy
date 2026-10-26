package application

import (
	"context"
	"sort"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func (application *LayoutApplication) List(
	ctx context.Context,
	resource layoutapi.ResourceKind,
) ([]layoutapi.Revision, error) {
	if !application.usable(ctx) || !resource.Valid() {
		return nil, layoutapi.ErrInvalid
	}
	revisions, err := application.store.List(ctx, resource)
	if err != nil {
		return nil, err
	}
	result := make([]layoutapi.Revision, len(revisions))
	ids := make(map[string]struct{}, len(revisions))
	for index, revision := range revisions {
		if revision.Resource != resource || revision.ID == "" ||
			revision.Version < 1 || !revision.State.Valid() {
			return nil, layoutapi.ErrUnavailable
		}
		if _, exists := ids[revision.ID]; exists {
			return nil, layoutapi.ErrUnavailable
		}
		ids[revision.ID] = struct{}{}
		result[index] = cloneRevision(revision)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].Version < result[right].Version
	})
	for index := 1; index < len(result); index++ {
		if result[index].Version == result[index-1].Version {
			return nil, layoutapi.ErrUnavailable
		}
	}
	return result, nil
}
