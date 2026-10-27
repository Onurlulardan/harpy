package main

import "testing"

func TestInspectAllowsRemainingTree(t *testing.T) {
	root := t.TempDir()
	mod := fixtureModulePath
	for _, file := range []struct{ path, body string }{
		{"cmd/api/main.go", goImports("os", mod+"/internal/composition/api")},
		{"cmd/worker/main.go", goImports("fmt", mod+"/internal/composition/worker")},
		{"cmd/migrator/main.go", goImports("context", mod+"/internal/composition/migrator")},
		{"cmd/operator/main.go", goImports("os", mod+"/internal/composition/operator")},
		{"core/cmd/api/main.go", goImports("os", mod+"/core/internal/composition/api")},
		{"core/cmd/worker/main.go", goImports("os", mod+"/core/internal/composition/worker")},
		{"core/cmd/migrator/main.go", goImports("os", mod+"/core/internal/composition/migrator")},
		{"core/cmd/operator/main.go", goImports("os", mod+"/core/internal/composition/operator")},
		{"cmd/setup/main.go", goImports(
			"os", "gopkg.in/yaml.v3",
			mod+"/internal/platform/config", mod+"/internal/platform/rustfs",
			mod+"/internal/modules/iam/api", mod+"/internal/modules/iam/api/systemresource",
			mod+"/internal/modules/iam/domain")},
		{"core/cmd/setup/ask.go", goImports(
			"os", mod+"/core/internal/platform/config",
			mod+"/core/internal/modules/tenancy/api",
			mod+"/core/internal/modules/tenancy/domain/value")},
		{"internal/platform/config/load.go", goImports("encoding/json")},
		{"internal/platform/postgres/pool.go", goImports("database/sql", "github.com/jackc/pgx/v5")},
		{"internal/platform/rustfs/endpoint.go", goImports("net/url")},
		{"internal/composition/api/wire.go", goImports(
			"context", mod+"/internal/modules/iam", mod+"/internal/modules/iam/application",
			mod+"/internal/modules/iam/adapters/postgres", mod+"/internal/platform/postgres")},
		{"internal/modules/iam/wire.go", goImports(
			mod+"/internal/modules/iam/application", mod+"/internal/modules/iam/adapters/postgres",
			mod+"/internal/modules/tenancy/api")},
		{"internal/modules/iam/api/contract.go", goImports("context", mod+"/internal/modules/iam/domain")},
		{"internal/modules/iam/application/create.go", goImports(
			"github.com/google/uuid", mod+"/internal/modules/iam/domain",
			mod+"/internal/modules/iam/domain/value", mod+"/internal/modules/tenancy/api/scope",
			mod+"/internal/modules/workflowengine/api")},
		{"internal/modules/iam/domain/user.go", goImports(
			"strings", "net/url", "github.com/google/uuid", "gopkg.in/yaml.v3",
			mod+"/internal/modules/tenancy/api")},
		{"internal/modules/iam/domain/value/id.go", goImports("strings")},
		{"internal/modules/iam/adapters/http/handle.go", goImports(
			"net/http", mod+"/internal/modules/iam/application", mod+"/internal/modules/iam/domain",
			mod+"/internal/modules/tenancy/api")},
		{"internal/modules/iam/adapters/postgres/store.go", goImports(
			"database/sql", "github.com/jackc/pgx/v5",
			mod+"/internal/modules/iam/application", mod+"/internal/modules/iam/domain")},
		{"internal/modules/documents/adapters/s3blob/client.go", goImports(
			"github.com/aws/aws-sdk-go-v2", "github.com/aws/aws-sdk-go-v2/service/s3",
			mod+"/internal/modules/iam/api")},
		{"internal/modules/iam/adapters/systemtransferartifact/store.go", goImports(
			"github.com/aws/aws-sdk-go-v2/service/s3", mod+"/internal/modules/iam/application")},
		{"core/internal/capabilities/artifactobject/adapters/s3object/client.go", goImports(
			"github.com/aws/aws-sdk-go-v2/service/s3")},
		{"internal/modules/work/application/run.go", goImports(
			mod+"/internal/modules/work/domain", mod+"/internal/modules/iam/api")},
		{"core/internal/modules/tenancy/application/open.go", goImports(
			mod+"/core/internal/modules/tenancy/domain", mod+"/core/internal/modules/iam/api")},
		{"internal/modules/formsetup/domain/name.go", goImports("strings")},
		{"internal/modules/iam/domain/email.go", goImports("strings")},
		{"internal/modules/iam/domain/claim.go", goImports("strings")},
		{"internal/modules/notifications/application/workflow_action.go", goImports("context")},
	} {
		writeFixture(t, root, file.path, file.body)
	}
	if violations := inspectFixture(t, root); len(violations) != 0 {
		t.Fatalf("allowed tree rejected: %v", violations)
	}
}
