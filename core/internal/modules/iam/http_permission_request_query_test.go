package iam

import (
	"net/http/httptest"
	"testing"
)

func TestPermissionRequestListRejectsAdHocStateFilter(t *testing.T) {
	request := httptest.NewRequest(
		"GET", "/api/permission-requests?states=PENDING", nil,
	)
	if _, valid := permissionRequestListInput(request); valid {
		t.Fatal("ad-hoc state filter bypassed canonical REST Query boundary")
	}
	request = httptest.NewRequest(
		"GET", "/api/permission-requests?limit=25", nil,
	)
	input, valid := permissionRequestListInput(request)
	if !valid || input.Limit != 25 {
		t.Fatalf("fixed list input = %#v / %v", input, valid)
	}
}
