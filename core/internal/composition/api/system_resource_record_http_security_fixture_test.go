package api

import (
	"context"
	"encoding/json"
	"testing"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

type recordHTTPSecurityPrepared struct{ resource systemresources.Resource }

func (prepared recordHTTPSecurityPrepared) Resource() (systemresources.Resource, error) {
	return prepared.resource, nil
}
func (recordHTTPSecurityPrepared) Contract() string {
	return "iam-system-resource-prepared-create.v1"
}
func (recordHTTPSecurityPrepared) Destroy() {}

type recordHTTPSecurityApplication struct {
	fields       map[systemresources.Resource][]systemResourceRecordHTTPField
	calls        map[string]systemresources.Resource
	createValues map[systemresources.ID]json.RawMessage
	updateValues map[systemresources.ID]json.RawMessage
	password     []byte
	createKey    string
	updateInput  systemresources.UpdateInput
	deleteInput  systemresources.DeleteInput
	view         systemResourceRecordHTTPView
}

func (app *recordHTTPSecurityApplication) Fields(
	_ context.Context, _ systemresources.Admission, resource systemresources.Resource,
) ([]systemResourceRecordHTTPField, error) {
	return append([]systemResourceRecordHTTPField(nil), app.fields[resource]...), nil
}

func (app *recordHTTPSecurityApplication) PrepareCreate(
	_ context.Context, input systemresources.CreateInput,
) (systemresources.PreparedCreate, error) {
	resource, _ := input.Resource()
	values, _ := input.Values()
	app.calls["prepare"], app.createValues = resource, recordHTTPSecurityValues(values)
	app.password, _ = input.Password()
	app.createKey, _ = input.IdempotencyKey()
	return recordHTTPSecurityPrepared{resource: resource}, nil
}

func (app *recordHTTPSecurityApplication) Create(
	_ context.Context, _ systemresources.Admission, prepared systemresources.PreparedCreate,
) (systemResourceRecordHTTPView, error) {
	resource, _ := prepared.Resource()
	app.calls["create"] = resource
	return app.view, nil
}

func (app *recordHTTPSecurityApplication) Get(
	_ context.Context, _ systemresources.Admission, input systemresources.GetInput,
) (systemResourceRecordHTTPView, error) {
	app.calls["get"], _ = input.Resource()
	return app.view, nil
}

func (app *recordHTTPSecurityApplication) Query(
	_ context.Context, _ systemresources.Admission, input systemresources.QueryInput,
) (systemResourceRecordHTTPQueryView, error) {
	app.calls["list"], _ = input.Resource()
	item, _ := json.Marshal(app.view.Data)
	return systemResourceRecordHTTPQueryView{Items: []json.RawMessage{item}}, nil
}

func (app *recordHTTPSecurityApplication) Update(
	_ context.Context, _ systemresources.Admission, input systemresources.UpdateInput,
) (systemResourceRecordHTTPView, error) {
	app.calls["update"], app.updateInput = recordHTTPSecurityUpdate(input)
	values, _ := input.Values()
	app.updateValues = recordHTTPSecurityValues(values)
	return app.view, nil
}

func (app *recordHTTPSecurityApplication) Delete(
	_ context.Context, _ systemresources.Admission, input systemresources.DeleteInput,
) error {
	app.calls["delete"], app.deleteInput = recordHTTPSecurityDelete(input)
	return nil
}

type recordHTTPSecurityAccessError struct{ err error }

func (access recordHTTPSecurityAccessError) Verify(
	context.Context, systemresources.LayoutAccessRequest,
) (systemresources.LayoutAccessDecision, error) {
	return systemresources.LayoutAccessDecision{}, access.err
}

func newRecordHTTPSecurityApplication(t *testing.T) *recordHTTPSecurityApplication {
	t.Helper()
	field := func(value string) systemresources.ID {
		id, err := systemresources.ParseID(value)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	name := func(id systemresources.ID, value string) systemResourceRecordHTTPField {
		return systemResourceRecordHTTPField{ID: id, Name: value, Editable: true}
	}
	fields := map[systemresources.Resource][]systemResourceRecordHTTPField{
		systemresources.Users: {
			name(field("9c65f0c0-1000-4000-8000-000000000001"), "name"),
		},
		systemresources.Roles: {
			name(field("9c65f0c0-1000-4000-8000-000000000002"), "name"),
			name(field("9c65f0c0-1000-4000-8000-000000000006"), "display_name"),
			name(field("9c65f0c0-1000-4000-8000-000000000003"), "scope_kind"),
			name(field("9c65f0c0-1000-4000-8000-000000000004"), "scope_id"),
			name(field("9c65f0c0-1000-4000-8000-000000000008"), "is_default"),
		},
		systemresources.Organizations: {
			name(field("9c65f0c0-1000-4000-8000-000000000005"), "name"),
			name(field("9c65f0c0-1000-4000-8000-000000000007"), "display_name"),
			name(field("9c65f0c0-1000-4000-8000-00000000000c"), "owner_ids"),
		},
	}
	data := map[string]json.RawMessage{
		"id":       json.RawMessage(`"` + recordHTTPTestID + `"`),
		"revision": json.RawMessage(`11`),
		"name":     json.RawMessage(`"visible"`),
	}
	return &recordHTTPSecurityApplication{
		fields: fields, calls: map[string]systemresources.Resource{},
		view: systemResourceRecordHTTPView{Data: data, ID: recordHTTPTestID, Revision: 11},
	}
}

func recordHTTPSecurityValues(
	values systemresources.FieldValues,
) map[systemresources.ID]json.RawMessage {
	result, _ := values.Values()
	return result
}

func recordHTTPSecurityUpdate(
	input systemresources.UpdateInput,
) (systemresources.Resource, systemresources.UpdateInput) {
	resource, _ := input.Resource()
	return resource, input
}

func recordHTTPSecurityDelete(
	input systemresources.DeleteInput,
) (systemresources.Resource, systemresources.DeleteInput) {
	resource, _ := input.Resource()
	return resource, input
}
