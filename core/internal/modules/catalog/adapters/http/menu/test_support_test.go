package menuhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/google/uuid"
)

type menuAccess struct{ operation string }

func (access *menuAccess) Verify(_ context.Context,
	request cataloghttp.AccessRequest,
) (cataloghttp.Access, error) {
	access.operation = request.Operation
	return cataloghttp.NewAccess(uuid.New(), uuid.New(), 1, 1, 1, time.Time{},
		request.Permission, request.Operation)
}

type menuApplication struct{ called string }

func (app *menuApplication) GetMenu(context.Context,
	cataloghttp.Access,
) (cataloghttp.Menu, error) {
	app.called = "get_menu"
	return cataloghttp.Menu{}, nil
}

func (app *menuApplication) UpdateMenuOverride(context.Context, cataloghttp.Access,
	cataloghttp.MenuOverrideInput,
) (cataloghttp.MenuOverrideReceipt, error) {
	app.called = "update_menu_override"
	return cataloghttp.MenuOverrideReceipt{}, nil
}

func (app *menuApplication) ReplaceMenu(context.Context, cataloghttp.Access,
	cataloghttp.MenuReplaceInput,
) (cataloghttp.Menu, error) {
	app.called = "replace_menu"
	return cataloghttp.Menu{}, nil
}

func (app *menuApplication) GetActiveMenuConfig(context.Context,
	cataloghttp.Access,
) (*cataloghttp.MenuConfig, error) {
	app.called = "get_active_menu_config"
	return nil, nil
}

func (app *menuApplication) CreateMenuConfig(context.Context, cataloghttp.Access,
	cataloghttp.MenuConfigWriteInput,
) (cataloghttp.MenuConfig, error) {
	app.called = "create_menu_config"
	return cataloghttp.MenuConfig{
		ID: "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41", Name: "Menu",
		Config: json.RawMessage(`{"groups":[],"ungrouped":[]}`), IsActive: true,
		CreatedBy: "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41",
		UpdatedBy: "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41",
	}, nil
}

func (app *menuApplication) UpdateMenuConfig(context.Context, cataloghttp.Access,
	uuid.UUID, cataloghttp.MenuConfigWriteInput,
) (cataloghttp.MenuConfig, error) {
	app.called = "update_menu_config"
	return cataloghttp.MenuConfig{
		ID: "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41", Name: "Menu",
		Config: json.RawMessage(`{"groups":[],"ungrouped":[]}`), IsActive: true,
		CreatedBy: "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41",
		UpdatedBy: "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41",
	}, nil
}

func exerciseMenu(expectedCall, expectedOperation string,
	selectRoute func(*Handler) Route,
) error {
	access := new(menuAccess)
	application := new(menuApplication)
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
	path := strings.ReplaceAll(strings.ReplaceAll(route.Pattern, "{item_id}", id.String()),
		"{id}", id.String())
	body, wantStatus := menuRouteBody(route.Method, route.Pattern)
	request := httptest.NewRequest(route.Method, path, strings.NewReader(body))
	request.SetPathValue("item_id", id.String())
	request.SetPathValue("id", id.String())
	if route.Method != http.MethodGet {
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", "catalog-test-key-0001")
	}
	response := httptest.NewRecorder()
	route.Handler.ServeHTTP(response, request)
	if response.Code != wantStatus || application.called != expectedCall ||
		access.operation != expectedOperation {
		return fmt.Errorf("menu route status=%d call=%s operation=%s",
			response.Code, application.called, access.operation)
	}
	return nil
}

func menuRouteBody(method, pattern string) (string, int) {
	switch {
	case method == http.MethodPatch:
		return `{"expected_revision":1,"visible":true}`, http.StatusOK
	case method == http.MethodPost:
		return `{"name":"Menu","config":{"groups":[],"ungrouped":[]}}`, http.StatusCreated
	case method == http.MethodPut && strings.Contains(pattern, "/configs/"):
		return `{"name":"Menu","config":{"groups":[],"ungrouped":[]},"is_active":true}`,
			http.StatusOK
	case method == http.MethodPut:
		return `{"expected_revision":1,"items":[]}`, http.StatusOK
	default:
		return "", http.StatusOK
	}
}
