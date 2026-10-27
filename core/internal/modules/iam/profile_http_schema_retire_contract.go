package iam

import (
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func profileRetirementChange(
	active profileapp.Schema,
	input ProfileSchemaRetire,
) (profileapp.SchemaChange, error) {
	fields := active.Fields()
	found := false
	for index, field := range fields {
		if field.ID() != input.FieldID {
			continue
		}
		retired, err := profileapp.RetiredField(field)
		if err != nil {
			return profileapp.SchemaChange{}, profileApplicationError(err)
		}
		fields[index], found = retired, true
	}
	if !found {
		return profileapp.SchemaChange{}, ErrProfileHTTPNotFound
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: input.SchemaID, Resource: input.Resource,
		Version: active.Version() + 1, State: profileapp.SchemaActive,
		Fields: fields,
	})
	if err != nil {
		return profileapp.SchemaChange{}, profileApplicationError(err)
	}
	return profileapp.SchemaChange{
		Schema: schema, ExpectedActiveSchemaID: active.ID(),
	}, nil
}

func profileRetireRequestSHA(input ProfileSchemaRetire) ([32]byte, error) {
	return identityRequestSHA(struct {
		SchemaID, ExpectedActiveSchemaID, FieldID [16]byte
		Resource                                  profileapp.ResourceKind
		ExpectedFieldRevision                     int64
	}{
		input.SchemaID, input.ExpectedActiveSchemaID, input.FieldID,
		input.Resource, input.ExpectedFieldRevision,
	})
}
