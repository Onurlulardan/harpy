package api

import (
	"context"
	"fmt"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

type canonicalSystemResourceRecordHTTPApplication struct {
	port systemresources.Port
}

func newCanonicalSystemResourceRecordHTTPApplication(
	port systemresources.Port,
) (*canonicalSystemResourceRecordHTTPApplication, error) {
	if port == nil {
		return nil, fmt.Errorf("canonical system resource application is required")
	}
	return &canonicalSystemResourceRecordHTTPApplication{port: port}, nil
}

func (application *canonicalSystemResourceRecordHTTPApplication) Fields(
	ctx context.Context,
	admission systemresources.Admission,
	resource systemresources.Resource,
) ([]systemResourceRecordHTTPField, error) {
	result, err := application.port.Catalog(ctx, admission, resource)
	if err != nil {
		return nil, err
	}
	presentation, err := result.Presentation()
	if err != nil {
		return nil, systemresources.ErrUnavailable
	}
	catalog, err := presentation.Catalog()
	if err != nil {
		return nil, systemresources.ErrUnavailable
	}
	descriptors, err := catalog.Descriptors()
	if err != nil {
		return nil, systemresources.ErrUnavailable
	}
	fields := make([]systemResourceRecordHTTPField, len(descriptors))
	for index, descriptor := range descriptors {
		id, idErr := descriptor.ID()
		name, nameErr := descriptor.PublicName()
		editable, editableErr := catalog.Editable(id)
		if idErr != nil || nameErr != nil || editableErr != nil {
			return nil, systemresources.ErrUnavailable
		}
		fields[index] = systemResourceRecordHTTPField{
			ID: id, Name: name, Editable: editable,
		}
	}
	return fields, nil
}

func (application *canonicalSystemResourceRecordHTTPApplication) PrepareCreate(
	ctx context.Context,
	input systemresources.CreateInput,
) (systemresources.PreparedCreate, error) {
	return application.port.PrepareCreate(ctx, input)
}

func (application *canonicalSystemResourceRecordHTTPApplication) Create(
	ctx context.Context,
	admission systemresources.Admission,
	prepared systemresources.PreparedCreate,
) (systemResourceRecordHTTPView, error) {
	result, err := application.port.Create(ctx, admission, prepared)
	if err != nil {
		return systemResourceRecordHTTPView{}, err
	}
	mutation, err := result.IAM()
	if err != nil {
		return systemResourceRecordHTTPView{}, systemresources.ErrUnavailable
	}
	read, err := mutation.Read()
	if err != nil {
		return systemResourceRecordHTTPView{}, systemresources.ErrUnavailable
	}
	return canonicalSystemResourceRecordHTTPView(read)
}

func (application *canonicalSystemResourceRecordHTTPApplication) Get(
	ctx context.Context,
	admission systemresources.Admission,
	input systemresources.GetInput,
) (systemResourceRecordHTTPView, error) {
	result, err := application.port.Get(ctx, admission, input)
	if err != nil {
		return systemResourceRecordHTTPView{}, err
	}
	read, err := result.IAM()
	if err != nil {
		return systemResourceRecordHTTPView{}, systemresources.ErrUnavailable
	}
	return canonicalSystemResourceRecordHTTPView(read)
}

func (application *canonicalSystemResourceRecordHTTPApplication) Query(
	ctx context.Context,
	admission systemresources.Admission,
	input systemresources.QueryInput,
) (systemResourceRecordHTTPQueryView, error) {
	result, err := application.port.Query(ctx, admission, input)
	if err != nil {
		return systemResourceRecordHTTPQueryView{}, err
	}
	query, err := result.IAM()
	if err != nil {
		return systemResourceRecordHTTPQueryView{}, systemresources.ErrUnavailable
	}
	items, hasMore, cursor, exactCount, err := projectSystemResourceQuery(query)
	if err != nil {
		return systemResourceRecordHTTPQueryView{}, err
	}
	return systemResourceRecordHTTPQueryView{
		Items: items, HasMore: hasMore, NextCursor: cursor, ExactCount: exactCount,
	}, nil
}

func (application *canonicalSystemResourceRecordHTTPApplication) Update(
	ctx context.Context,
	admission systemresources.Admission,
	input systemresources.UpdateInput,
) (systemResourceRecordHTTPView, error) {
	result, err := application.port.Update(ctx, admission, input)
	if err != nil {
		return systemResourceRecordHTTPView{}, err
	}
	mutation, err := result.IAM()
	if err != nil {
		return systemResourceRecordHTTPView{}, systemresources.ErrUnavailable
	}
	read, err := mutation.Read()
	if err != nil {
		return systemResourceRecordHTTPView{}, systemresources.ErrUnavailable
	}
	return canonicalSystemResourceRecordHTTPView(read)
}

func (application *canonicalSystemResourceRecordHTTPApplication) Delete(
	ctx context.Context,
	admission systemresources.Admission,
	input systemresources.DeleteInput,
) error {
	_, err := application.port.Delete(ctx, admission, input)
	return err
}

func canonicalSystemResourceRecordHTTPView(
	read systemresources.ReadRecord,
) (systemResourceRecordHTTPView, error) {
	data, revision, id, err := projectSystemResourceRead(read)
	if err != nil {
		return systemResourceRecordHTTPView{}, err
	}
	return systemResourceRecordHTTPView{Data: data, ID: id, Revision: revision}, nil
}
