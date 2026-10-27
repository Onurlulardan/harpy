package api

import (
	"context"
	"encoding/json"
	"strings"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func (handler *systemResourceRecordHTTPHandler) mintRecordCreate(
	ctx context.Context,
	admission systemresources.Admission,
	resource systemresources.Resource,
	document map[string]json.RawMessage,
	idempotencyKey string,
) (systemresources.PreparedCreate, error) {
	if err := normalizeSystemResourceCreate(
		admission, resource, document,
	); err != nil {
		return nil, err
	}
	password, err := systemResourceRecordPassword(resource, document)
	if err != nil {
		return nil, err
	}
	defer clear(password)
	values, err := handler.mintRecordFieldValues(
		ctx, admission, resource, document,
	)
	if err != nil {
		return nil, err
	}
	input, err := systemresources.NewCreateInput(
		resource, values, password, idempotencyKey,
	)
	if err != nil {
		return nil, systemresources.ErrContract
	}
	return handler.application.PrepareCreate(ctx, input)
}

func (handler *systemResourceRecordHTTPHandler) mintRecordUpdate(
	ctx context.Context,
	admission systemresources.Admission,
	resource systemresources.Resource,
	id systemresources.ID,
	expectedRevision int64,
	document map[string]json.RawMessage,
	idempotencyKey string,
) (systemresources.UpdateInput, error) {
	if _, forbidden := document["password"]; forbidden {
		return systemresources.UpdateInput{}, systemresources.ErrContract
	}
	normalizeSystemResourceNullableRelations(resource, document)
	if resource == systemresources.Organizations {
		if err := normalizeSystemOrganizationOwners(document); err != nil {
			return systemresources.UpdateInput{}, err
		}
	}
	values, err := handler.mintRecordFieldValues(
		ctx, admission, resource, document,
	)
	if err != nil {
		return systemresources.UpdateInput{}, err
	}
	input, err := systemresources.NewUpdateInput(
		resource, id, expectedRevision, values, idempotencyKey,
	)
	if err != nil {
		return systemresources.UpdateInput{}, systemresources.ErrContract
	}
	return input, nil
}

func mintRecordDelete(
	resource systemresources.Resource,
	id systemresources.ID,
	expectedRevision int64,
	idempotencyKey string,
) (systemresources.DeleteInput, error) {
	input, err := systemresources.NewDeleteInput(
		resource, id, expectedRevision, idempotencyKey,
	)
	if err != nil {
		return systemresources.DeleteInput{}, systemresources.ErrContract
	}
	return input, nil
}

func systemResourceRecordPassword(
	resource systemresources.Resource,
	document map[string]json.RawMessage,
) ([]byte, error) {
	raw, present := document["password"]
	if resource != systemresources.Users {
		if present {
			return nil, systemresources.ErrContract
		}
		return nil, nil
	}
	delete(document, "password")
	var value string
	if !present || json.Unmarshal(raw, &value) != nil || strings.TrimSpace(value) == "" {
		clear(raw)
		return nil, systemresources.ErrContract
	}
	clear(raw)
	return []byte(value), nil
}
