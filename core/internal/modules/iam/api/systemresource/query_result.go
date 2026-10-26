package systemresource

import (
	"bytes"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

const (
	maximumQueryItems     = 500
	maximumQueryItemBytes = 1 << 20
	maximumCursorBytes    = 16 << 10
)

type QueryResult struct {
	items      []json.RawMessage
	hasMore    bool
	nextCursor string
	exactCount *int64
	catalog    FieldCatalog
	relations  []RelationProjection
	validated  bool
}

func NewQueryResult(
	items []json.RawMessage,
	hasMore bool,
	nextCursor string,
	exactCount *int64,
	catalog FieldCatalog,
) (QueryResult, error) {
	resource, err := catalog.Resource()
	if err != nil {
		return QueryResult{}, ErrContract
	}
	relations := make([]RelationProjection, len(items))
	for index := range relations {
		relations[index], err = EmptyRelationProjection(resource)
		if err != nil {
			return QueryResult{}, ErrContract
		}
	}
	return NewQueryResultWithRelations(
		items, hasMore, nextCursor, exactCount, catalog, relations,
	)
}

func NewQueryResultWithRelations(
	items []json.RawMessage,
	hasMore bool,
	nextCursor string,
	exactCount *int64,
	catalog FieldCatalog,
	relations []RelationProjection,
) (QueryResult, error) {
	owned, err := canonicalQueryItems(items)
	if err != nil {
		return QueryResult{}, err
	}
	result := QueryResult{
		items: owned, hasMore: hasMore, nextCursor: nextCursor,
		exactCount: cloneCount(exactCount), catalog: catalog,
		relations: append([]RelationProjection(nil), relations...), validated: true,
	}
	if result.Validate() != nil {
		return QueryResult{}, ErrContract
	}
	return result, nil
}

func canonicalQueryItems(items []json.RawMessage) ([]json.RawMessage, error) {
	if items == nil || len(items) > maximumQueryItems {
		return nil, ErrContract
	}
	result := make([]json.RawMessage, len(items))
	for index, item := range items {
		canonical, err := canonicaljsonv1.CanonicalizeObject(item, maximumQueryItemBytes)
		if err != nil {
			return nil, ErrContract
		}
		result[index] = canonical
	}
	return result, nil
}

func (result QueryResult) Validate() error {
	resource, resourceErr := result.catalog.Resource()
	if !result.validated || result.items == nil || len(result.items) > maximumQueryItems ||
		result.catalog.Validate() != nil || result.hasMore != (result.nextCursor != "") ||
		len(result.nextCursor) > maximumCursorBytes ||
		result.exactCount != nil && *result.exactCount < 0 ||
		resourceErr != nil || len(result.relations) != len(result.items) {
		return ErrContract
	}
	for index, item := range result.items {
		canonical, err := canonicaljsonv1.CanonicalizeObject(item, maximumQueryItemBytes)
		relationResource, _, relationErr := result.relations[index].Values()
		if err != nil || !bytes.Equal(canonical, item) ||
			relationErr != nil || relationResource != resource {
			return ErrContract
		}
	}
	return nil
}

func cloneCount(value *int64) *int64 {
	if value == nil {
		return nil
	}
	owned := *value
	return &owned
}

func (result QueryResult) Items() ([]json.RawMessage, error) {
	if result.Validate() != nil {
		return nil, ErrContract
	}
	items := make([]json.RawMessage, len(result.items))
	for index, item := range result.items {
		items[index] = append(json.RawMessage(nil), item...)
	}
	return items, nil
}

func (result QueryResult) Pagination() (bool, string, *int64, error) {
	if result.Validate() != nil {
		return false, "", nil, ErrContract
	}
	return result.hasMore, result.nextCursor, cloneCount(result.exactCount), nil
}

func (result QueryResult) Catalog() (FieldCatalog, error) {
	if result.Validate() != nil {
		return FieldCatalog{}, ErrContract
	}
	return result.catalog, nil
}

func (result QueryResult) Relations() ([]RelationProjection, error) {
	if result.Validate() != nil {
		return nil, ErrContract
	}
	return append([]RelationProjection(nil), result.relations...), nil
}
