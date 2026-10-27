//go:build integration

package integrationtest

import (
	"encoding/json"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	iampayload "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/catalogpayload"
)

const (
	iamCatalogV1OperationID = "018f0000-0004-7000-8000-000000000006"
	iamCatalogV2OperationID = "018f0000-0004-7000-8000-000000000010"
	iamCatalogV2ReceiptID   = "018f0000-0004-7000-8000-000000000011"
	iamCatalogV2ReleaseID   = "018f0000-0004-7000-8000-000000000012"
	iamCatalogRetryOpID     = "018f0000-0004-7000-8000-000000000017"
	iamCatalogRetryReceipt  = "018f0000-0004-7000-8000-000000000018"
)

func iamCatalogDocumentV2(t *testing.T) []byte {
	t.Helper()
	payload := iampayload.Payload{
		ContractVersion: catalogpackagev1.ContractIAM,
		ComponentID:     iamCatalogComponentID,
		ModuleSlug:      "project",
		SourceContract:  "harpy-v1.enterprise-seed.project",
		Resources: []iampayload.Resource{
			{
				Slug: "module.project", Kind: "module",
				Name: "Project", Description: "Module",
			},
			{
				Slug: "form.project-task", Kind: "form",
				Name: "Task v2", Description: "Form v2",
			},
			{
				Slug: "workflow.project-task-review", Kind: "workflow",
				Name: "Task Review", Description: "Workflow",
			},
		},
		Grants: []iampayload.Grant{
			{
				ResourceSlug:    "module.project",
				RoleName:        "Project Administrator",
				RoleDescription: "Project administrator",
				Actions:         []string{"read"},
			},
			{
				ResourceSlug:    "form.project-task",
				RoleName:        "Project Administrator",
				RoleDescription: "Project administrator",
				Actions:         []string{"read", "write"},
			},
			{
				ResourceSlug:    "workflow.project-task-review",
				RoleName:        "Project Administrator",
				RoleDescription: "Project administrator",
				Actions:         []string{"read", "write"},
			},
		},
	}
	document, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 4<<20)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}
