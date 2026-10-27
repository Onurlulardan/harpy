package work

import (
	"context"
	"encoding/json"

	workpostgres "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/postgres"
	workquery "github.com/Onurlulardan/harpy/core/internal/modules/work/application/restquery"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *PlatformHTTPRuntime) List(ctx context.Context,
	input HTTPListInput,
) (result json.RawMessage, err error) {
	if !validHTTPAccess(input.Access, input.Operation, input.Resource, input.Compatibility) {
		return nil, ErrHTTPForbidden
	}
	values, err := runtime.resolve(ctx, input.Access)
	if err != nil {
		return nil, ErrHTTPInvalid
	}
	if !input.Compatibility {
		return runtime.queryList(ctx, input, values)
	}
	options, err := legacyReadOptions(input)
	if err != nil {
		return nil, err
	}
	options.Compatibility = input.Compatibility
	target, parent, err := readIDs(input.Resource, input.IDs)
	if err != nil {
		return nil, err
	}
	formsScope := compatibilityReadNeedsForms(input.Resource)
	if formsScope {
		values, err = runtime.resolveForms(ctx, input.Access, values)
		if err != nil {
			return nil, err
		}
	}
	limit := input.Query.Page().Size()
	err = runtime.compatibilitySnapshot(ctx, input.Access, values, input.Access.Authority(),
		formsScope,
		func(operation context.Context, scope platformtx.TenantScope) error {
			if input.Resource == "available-users" {
				items, readErr := runtime.principals.AvailableUsers(operation, scope,
					input.LegacyQuery["search"], limit)
				if readErr == nil {
					result, readErr = listDocument(items, limit, input.Compatibility)
				}
				return readErr
			}
			items, readErr := runtime.reader.Read(operation,
				workpostgres.BridgeScope(scope), input.Resource, target, parent, limit, options)
			if readErr == nil && input.Resource == "tasks" {
				items, readErr = runtime.enrichLegacyTaskRead(operation,
					workpostgres.BridgeScope(scope), items, false)
			}
			if readErr == nil {
				result, readErr = listDocument(items, limit, input.Compatibility)
			}
			return readErr
		})
	return result, err
}

func (runtime *PlatformHTTPRuntime) queryList(ctx context.Context,
	input HTTPListInput, values workGenerations,
) (result json.RawMessage, err error) {
	err = runtime.snapshot(ctx, input.Access, values, input.Access.Authority(),
		func(operation context.Context, scope platformtx.TenantScope) error {
			evidence, queryErr := workquery.NewEvidence(input.Access.PrincipalID(),
				input.Access.TenantID(), input.Access.TokenVersion(),
				input.Access.AuthorizationVersion(), input.Resource, input.Operation, input.IDs)
			if queryErr != nil {
				return queryErr
			}
			queryResult, queryErr := runtime.query.Query(operation, scope, evidence, input.Query)
			if queryErr != nil {
				return queryErr
			}
			result, queryErr = workquery.Document(queryResult)
			return queryErr
		})
	return result, err
}

func (runtime *PlatformHTTPRuntime) Get(ctx context.Context,
	input HTTPGetInput,
) (result json.RawMessage, err error) {
	if !validHTTPAccess(input.Access, input.Operation, input.Resource, input.Compatibility) {
		return nil, ErrHTTPForbidden
	}
	values, err := runtime.resolve(ctx, input.Access)
	if err != nil {
		return nil, err
	}
	formsScope := input.Compatibility && compatibilityReadNeedsForms(input.Resource)
	if formsScope {
		values, err = runtime.resolveForms(ctx, input.Access, values)
		if err != nil {
			return nil, err
		}
	}
	resource := input.Resource
	if resource == "my-role" {
		resource = "boards"
	}
	target, parent, err := readIDs(resource, input.IDs)
	if err != nil {
		return nil, err
	}
	err = runtime.compatibilitySnapshot(ctx, input.Access, values, input.Access.Authority(),
		formsScope,
		func(operation context.Context, scope platformtx.TenantScope) error {
			if input.Compatibility && input.Resource == "boards" {
				items, readErr := runtime.reader.ReadBoardDetail(operation,
					workpostgres.BridgeScope(scope), target)
				if readErr == nil {
					items, readErr = runtime.enrichLegacyTaskRead(operation,
						workpostgres.BridgeScope(scope), items, true)
				}
				if readErr == nil {
					result, readErr = getDocument(items, input.Resource, true)
				}
				return readErr
			}
			items, readErr := runtime.reader.Read(operation,
				workpostgres.BridgeScope(scope), resource, target, parent, 1,
				workpostgres.ReadOptions{Compatibility: input.Compatibility})
			if readErr == nil && input.Compatibility && input.Resource == "tasks" {
				items, readErr = runtime.enrichLegacyTaskRead(operation,
					workpostgres.BridgeScope(scope), items, false)
			}
			if readErr == nil {
				result, readErr = getDocument(items, input.Resource, input.Compatibility)
			}
			return readErr
		})
	return result, err
}

func compatibilityReadNeedsForms(resource string) bool {
	return resource == "boards" || resource == "tasks"
}

func readIDs(resource string, ids map[string]string) (ports.ID, ports.ID, error) {
	var target, parent ports.ID
	var err error
	keys := map[string]string{"boards": "board_id", "tasks": "task_id",
		"calendar": "event_id", "comments": "comment_id"}
	if raw := ids[keys[resource]]; raw != "" {
		target, err = ports.ParseID(raw)
		if err != nil {
			return target, parent, ErrHTTPInvalid
		}
	}
	parentKey := map[string]string{"comments": "task_id", "memberships": "board_id"}[resource]
	if raw := ids[parentKey]; raw != "" {
		parent, err = ports.ParseID(raw)
		if err != nil {
			return target, parent, ErrHTTPInvalid
		}
	}
	return target, parent, nil
}
