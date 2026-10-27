package iam

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type iamPortalAuthorityContract struct {
	Authorities []struct {
		ID string `json:"id"`
	} `json:"authorities"`
}

type iamPortalOperationContract struct {
	Operations []struct {
		Source      string `json:"source"`
		AuthorityID string `json:"authority_id"`
	} `json:"operations"`
}

func TestPortalSetupStartContractsUseCanonicalAuthorityAndPermission(t *testing.T) {
	root := iamRepositoryRoot(t)
	document := iamReadOpenAPIPathDocument(
		t, root, "iam-lifecycle-account-v1.paths.json",
	)
	raw := document.PathItems["portalSetupToken"]["post"]
	var operation struct {
		Permission  json.RawMessage `json:"x-harpy-permission"`
		AuthorityID string          `json:"x-harpy-authority"`
	}
	if err := json.Unmarshal(raw, &operation); err != nil {
		t.Fatal(err)
	}
	if got := string(operation.Permission); got != `"iam.users.update"` {
		t.Fatalf("portal setup start permission = %s, want iam.users.update", got)
	}
	if operation.AuthorityID != "iam.portal_setup.start" {
		t.Fatalf(
			"portal setup start authority = %q, want iam.portal_setup.start",
			operation.AuthorityID,
		)
	}

	authorityWire, err := os.ReadFile(filepath.Join(
		root, "contracts", "authority", "iam_invitation_permission.json",
	))
	if err != nil {
		t.Fatal(err)
	}
	var authority iamPortalAuthorityContract
	if err = json.Unmarshal(authorityWire, &authority); err != nil {
		t.Fatal(err)
	}
	foundStart := false
	for _, entry := range authority.Authorities {
		switch entry.ID {
		case "iam.portal_setup.start":
			foundStart = true
		case "iam.portal_setup.create":
			t.Fatal("legacy iam.portal_setup.create authority is still published")
		}
	}
	if !foundStart {
		t.Fatal("canonical iam.portal_setup.start authority is not published")
	}

	operationWire, err := os.ReadFile(filepath.Join(
		root, "contracts", "operations", "iam_application_lifecycle.json",
	))
	if err != nil {
		t.Fatal(err)
	}
	var operations iamPortalOperationContract
	if err = json.Unmarshal(operationWire, &operations); err != nil {
		t.Fatal(err)
	}
	const source = "core/internal/modules/iam/application/lifecycle/create_portal_setup.go"
	for _, entry := range operations.Operations {
		if entry.Source == source {
			if entry.AuthorityID != "iam.portal_setup.start" {
				t.Fatalf(
					"%s authority = %q, want iam.portal_setup.start",
					source, entry.AuthorityID,
				)
			}
			return
		}
	}
	t.Fatalf("operation manifest does not include %s", source)
}
