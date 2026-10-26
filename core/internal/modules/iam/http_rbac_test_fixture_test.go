package iam

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

const rbacHTTPTestKey = "rbac-http-test-0001"

type rbacHTTPTestFixture struct {
	handler *RBACHTTP
	app     *rbacHTTPTestApplication
	token   string
	tenant  uuid.UUID
}

func newRBACHTTPTestFixture(t *testing.T) *rbacHTTPTestFixture {
	t.Helper()
	auth := newAuthHTTPTestFixture(t)
	app := &rbacHTTPTestApplication{}
	handler, err := newRBACHTTP(
		auth.security, app, func(error) {}, func() time.Time { return auth.now },
	)
	if err != nil {
		t.Fatal(err)
	}
	token, err := auth.security.accessSigner.Sign(auth.claims)
	if err != nil {
		t.Fatal(err)
	}
	return &rbacHTTPTestFixture{
		handler: handler, app: app, token: token, tenant: auth.claims.Tenant(),
	}
}

func (fixture *rbacHTTPTestFixture) request(
	method string,
	path string,
	body string,
) *http.Request {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Authorization", "Bearer "+fixture.token)
	request.Header.Set(authTenantIDHeader, fixture.tenant.String())
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	return request
}

func testRBACResourceView() RBACResourceView {
	return RBACResourceView{
		ID: uuid.New().String(), Name: "Users", Slug: "iam.users",
		OwnerDomain: "iam", Kind: "system_resource", State: "active", Revision: 1,
	}
}

func testRBACActionView() RBACActionView {
	return RBACActionView{
		ID: uuid.New().String(), Name: "Read", Slug: "read",
		State: "active", Revision: 1,
	}
}

func testRBACPermissionView() RBACPermissionView {
	userID := uuid.New().String()
	resource := testRBACResourceView()
	return RBACPermissionView{
		ID: uuid.New().String(), Target: "USER", ResourceID: resource.ID,
		UserID: &userID, Resource: resource,
		Actions: []RBACActionView{testRBACActionView()}, State: "active", Revision: 1,
	}
}
