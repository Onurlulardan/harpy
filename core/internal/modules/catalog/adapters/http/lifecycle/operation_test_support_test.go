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

type operationAccess struct{ operation string }

func (access *operationAccess) Verify(_ context.Context,
	request cataloghttp.AccessRequest,
) (cataloghttp.Access, error) {
	access.operation = request.Operation
	return cataloghttp.NewAccess(uuid.New(), uuid.New(), 1, 1, 1, time.Time{},
		request.Permission, request.Operation)
}

type operationApplication struct {
	*installationApplication
	called string
}

func (app *operationApplication) GetOperation(context.Context, cataloghttp.Access,
	uuid.UUID,
) (cataloghttp.Operation, error) {
	app.called = "get_operation"
	return cataloghttp.Operation{}, nil
}

func (app *operationApplication) CancelOperation(context.Context, cataloghttp.Access,
	uuid.UUID, string,
) (cataloghttp.Operation, error) {
	app.called = "cancel_operation"
	return cataloghttp.Operation{}, nil
}

func (app *operationApplication) RetryOperation(context.Context, cataloghttp.Access,
	uuid.UUID, string,
) (cataloghttp.Operation, error) {
	app.called = "retry_operation"
	return cataloghttp.Operation{}, nil
}

func exerciseOperation(expectedCall, expectedOperation string,
	selectRoute func(*Handler) Route,
) error {
	access := new(operationAccess)
	application := &operationApplication{installationApplication: new(installationApplication)}
	runtime, err := cataloghttp.NewRuntime(access, func(error) {}, func(*http.Request) string {
		return "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41"
	})
	if err != nil {
		return err
	}
	handler := New(runtime, application)
	route := selectRoute(handler)
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	path := strings.ReplaceAll(route.Pattern, "{operation_id}", id.String())
	request := httptest.NewRequest(route.Method, path, nil)
	request.SetPathValue("operation_id", id.String())
	if route.Method == http.MethodPost {
		request.Header.Set("Idempotency-Key", "catalog-test-key-0001")
	}
	response := httptest.NewRecorder()
	route.Handler.ServeHTTP(response, request)
	wantStatus := http.StatusOK
	if route.Method == http.MethodPost {
		wantStatus = http.StatusAccepted
	}
	if response.Code != wantStatus || application.called != expectedCall ||
		access.operation != expectedOperation {
		return fmt.Errorf("operation route status=%d call=%s operation=%s",
			response.Code, application.called, access.operation)
	}
	return nil
}
