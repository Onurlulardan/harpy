package lifecyclehttp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/google/uuid"
)

type installationAccess struct {
	operation, purpose string
	permission         cataloghttp.Permission
}

func (access *installationAccess) Verify(_ context.Context,
	request cataloghttp.AccessRequest,
) (cataloghttp.Access, error) {
	access.operation, access.purpose = request.Operation, request.PurposeTarget
	access.permission = request.Permission
	return cataloghttp.NewAccess(uuid.New(), uuid.New(), 1, 1, 1, time.Time{},
		request.Permission, request.Operation)
}

type installationApplication struct {
	called, key, expectedPlan string
}

func (app *installationApplication) ListInstallations(context.Context, cataloghttp.Access,
	int, *uuid.UUID,
) ([]cataloghttp.Installation, *uuid.UUID, error) {
	app.called = "list_installations"
	return []cataloghttp.Installation{}, nil, nil
}

func (app *installationApplication) Install(context.Context, cataloghttp.Access,
	uuid.UUID, string, string,
) (cataloghttp.Operation, error) {
	app.called = "install"
	return cataloghttp.Operation{}, nil
}

func (app *installationApplication) GetInstallation(context.Context, cataloghttp.Access,
	uuid.UUID,
) (cataloghttp.Installation, error) {
	app.called = "get_installation"
	return cataloghttp.Installation{}, nil
}

func (app *installationApplication) Upgrade(context.Context, cataloghttp.Access,
	uuid.UUID, string, string,
) (cataloghttp.Operation, error) {
	app.called = "upgrade"
	return cataloghttp.Operation{}, nil
}

func (app *installationApplication) Disable(context.Context, cataloghttp.Access,
	uuid.UUID, string,
) (cataloghttp.Installation, error) {
	app.called = "disable"
	return cataloghttp.Installation{}, nil
}

func (app *installationApplication) Enable(context.Context, cataloghttp.Access,
	uuid.UUID, string,
) (cataloghttp.Installation, error) {
	app.called = "enable"
	return cataloghttp.Installation{}, nil
}

func (app *installationApplication) Restore(context.Context, cataloghttp.Access,
	uuid.UUID, string,
) (cataloghttp.Operation, error) {
	app.called = "restore"
	return cataloghttp.Operation{}, nil
}

func (app *installationApplication) Retire(context.Context, cataloghttp.Access,
	uuid.UUID, string,
) (cataloghttp.Operation, error) {
	app.called = "retire"
	return cataloghttp.Operation{}, nil
}

func (app *installationApplication) Purge(context.Context, cataloghttp.Access,
	cataloghttp.PurgeInput,
) (cataloghttp.Operation, error) {
	app.called = "purge"
	return cataloghttp.Operation{}, nil
}

func (app *installationApplication) ApprovePurge(context.Context, cataloghttp.Access,
	cataloghttp.PurgeApprovalInput,
) (cataloghttp.PurgeApproval, error) {
	app.called = "approve_purge"
	return cataloghttp.PurgeApproval{}, nil
}

func (app *installationApplication) GetOperation(context.Context, cataloghttp.Access,
	uuid.UUID,
) (cataloghttp.Operation, error) {
	app.called = "get_operation"
	return cataloghttp.Operation{}, nil
}

func (app *installationApplication) CancelOperation(context.Context, cataloghttp.Access,
	uuid.UUID, string,
) (cataloghttp.Operation, error) {
	app.called = "cancel_operation"
	return cataloghttp.Operation{}, nil
}

func (app *installationApplication) RetryOperation(context.Context, cataloghttp.Access,
	uuid.UUID, string,
) (cataloghttp.Operation, error) {
	app.called = "retry_operation"
	return cataloghttp.Operation{}, nil
}

func exerciseInstallation(expectedCall, expectedOperation string,
	selectRoute func(*Handler) Route,
) error {
	access := new(installationAccess)
	application := new(installationApplication)
	runtime, err := cataloghttp.NewRuntime(access, func(error) {}, func(*http.Request) string {
		return "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41"
	})
	if err != nil {
		return err
	}
	handler := New(runtime, application)
	route := selectRoute(handler)
	installationID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	path := strings.ReplaceAll(route.Pattern, "{installation_id}", installationID.String())
	body, err := installationBody(expectedCall)
	if err != nil {
		return err
	}
	var request *http.Request
	if body == nil {
		request = httptest.NewRequest(route.Method, path, nil)
	} else {
		request = httptest.NewRequest(route.Method, path, body)
	}
	request.SetPathValue("installation_id", installationID.String())
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if route.Method == http.MethodPost {
		request.Header.Set("Idempotency-Key", "catalog-test-key-0001")
	}
	response := httptest.NewRecorder()
	route.Handler.ServeHTTP(response, request)
	wantStatus := installationStatus(expectedCall)
	if response.Code != wantStatus || application.called != expectedCall ||
		access.operation != expectedOperation || expectedCall == "purge" && access.purpose == "" {
		return fmt.Errorf("installation route status=%d call=%s operation=%s purpose=%s",
			response.Code, application.called, access.operation, access.purpose)
	}
	return nil
}
