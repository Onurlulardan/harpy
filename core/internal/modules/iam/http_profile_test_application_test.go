package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

type profileHTTPTestCall struct {
	operation string
	claims    accesstoken.Claims
	input     any
}

type profileHTTPTestApplication struct {
	calls    []profileHTTPTestCall
	results  map[string]any
	failures map[string]error
}

func newProfileHTTPTestApplication() *profileHTTPTestApplication {
	return &profileHTTPTestApplication{
		results: make(map[string]any), failures: make(map[string]error),
	}
}

func (application *profileHTTPTestApplication) invoke(
	operation string, claims accesstoken.Claims, input any,
) (any, error) {
	application.calls = append(application.calls, profileHTTPTestCall{
		operation: operation, claims: claims, input: input,
	})
	return application.results[operation], application.failures[operation]
}

func (application *profileHTTPTestApplication) GetSelf(
	_ context.Context, claims accesstoken.Claims,
) (ProfileSelfView, error) {
	result, err := application.invoke("GetSelf", claims, nil)
	view, _ := result.(ProfileSelfView)
	return view, err
}

func (application *profileHTTPTestApplication) UpdateSelf(
	_ context.Context, claims accesstoken.Claims, input ProfileSelfUpdate,
) (ProfileSelfView, error) {
	result, err := application.invoke("UpdateSelf", claims, input)
	view, _ := result.(ProfileSelfView)
	return view, err
}

func (application *profileHTTPTestApplication) GetActiveSchema(
	_ context.Context, claims accesstoken.Claims, resource profileapp.ResourceKind,
) (ProfileSchemaView, error) {
	result, err := application.invoke("GetActiveSchema", claims, resource)
	view, _ := result.(ProfileSchemaView)
	return view, err
}

func (application *profileHTTPTestApplication) GetSchema(
	_ context.Context, claims accesstoken.Claims,
	resource profileapp.ResourceKind, schemaID uuid.UUID,
) (ProfileSchemaView, error) {
	result, err := application.invoke("GetSchema", claims, struct {
		resource profileapp.ResourceKind
		id       uuid.UUID
	}{resource, schemaID})
	view, _ := result.(ProfileSchemaView)
	return view, err
}

func (application *profileHTTPTestApplication) PublishSchema(
	_ context.Context, claims accesstoken.Claims, input ProfileSchemaPublish,
) (ProfileSchemaMutationView, error) {
	result, err := application.invoke("PublishSchema", claims, input)
	view, _ := result.(ProfileSchemaMutationView)
	return view, err
}

func (application *profileHTTPTestApplication) RetireField(
	_ context.Context, claims accesstoken.Claims, input ProfileSchemaRetire,
) (ProfileSchemaMutationView, error) {
	result, err := application.invoke("RetireField", claims, input)
	view, _ := result.(ProfileSchemaMutationView)
	return view, err
}

func (application *profileHTTPTestApplication) RecoverBackfill(
	_ context.Context, claims accesstoken.Claims, input ProfileBackfillRecoveryInput,
) (ProfileBackfillRecoveryView, error) {
	result, err := application.invoke("RecoverBackfill", claims, input)
	view, _ := result.(ProfileBackfillRecoveryView)
	return view, err
}
