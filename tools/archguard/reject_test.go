package main

import "testing"

func TestInspectRejectsForbiddenImports(t *testing.T) {
	mod := modulePath
	tests := []struct{ name, source, imported, want string }{
		{"command platform", "cmd/api/main.go", mod + "/internal/platform/config", "matching composition entry"},
		{"command other profile", "cmd/worker/main.go", mod + "/internal/composition/api", "matching composition entry"},
		{"command third party", "cmd/migrator/main.go", "github.com/google/uuid", "matching composition entry"},
		{"command module", "cmd/operator/main_test.go", mod + "/internal/modules/iam/api", "matching composition entry"},
		{"flat command core entry", "cmd/api/main.go", mod + "/core/internal/composition/api", "matching composition entry"},
		{"core command flat entry", "core/cmd/api/main.go", mod + "/internal/composition/api", "matching composition entry"},
		{"core command platform", "core/cmd/worker/main.go", mod + "/core/internal/platform/config", "matching composition entry"},
		{"setup application", "cmd/setup/main.go", mod + "/internal/modules/iam/application", "application or adapter"},
		{"setup application child", "cmd/setup/main.go", mod + "/internal/modules/iam/application/login", "application or adapter"},
		{"setup adapters", "cmd/setup/main.go", mod + "/internal/modules/tenancy/adapters/postgres", "application or adapter"},
		{"setup singular adapter", "core/cmd/setup/main.go", mod + "/core/internal/modules/iam/adapter/http", "application or adapter"},
		{"platform module api", "internal/platform/http/edge.go", mod + "/internal/modules/iam/api", "platform cannot import modules"},
		{"platform module root", "core/internal/platform/postgres/pool.go", mod + "/core/internal/modules/tenancy", "platform cannot import modules"},
		{"domain http", "internal/modules/iam/domain/user.go", "net/http", "domain cannot import net/http"},
		{"domain httptest", "internal/modules/iam/domain/user.go", "net/http/httptest", "domain cannot import net/http"},
		{"domain sql", "internal/modules/work/domain/item.go", "database/sql", "domain cannot import database/sql"},
		{"domain driver", "internal/modules/work/domain/item.go", "database/sql/driver", "domain cannot import database/sql"},
		{"domain pgx", "internal/modules/billing/domain/money.go", "github.com/jackc/pgx/v5", "domain cannot import pgx"},
		{"domain pgtype", "internal/modules/billing/domain/money.go", "github.com/jackc/pgx/v5/pgtype", "domain cannot import pgx"},
		{"domain aws", "internal/modules/documents/domain/object.go", "github.com/aws/aws-sdk-go-v2/service/s3", "cloud SDKs"},
		{"domain gcp", "internal/modules/documents/domain/object.go", "cloud.google.com/go/storage", "cloud SDKs"},
		{"domain azure", "internal/modules/documents/domain/object.go", "github.com/Azure/azure-sdk-for-go/sdk/azblob", "cloud SDKs"},
		{"domain google api", "internal/modules/documents/domain/object.go", "google.golang.org/api/compute/v1", "cloud SDKs"},
		{"application adapter", "internal/modules/iam/application/create_test.go", mod + "/internal/modules/iam/adapters/postgres", "dependency direction"},
		{"domain application", "internal/modules/iam/domain/user.go", mod + "/internal/modules/iam/application", "dependency direction"},
		{"domain adapter", "internal/modules/catalog/domain/item.go", mod + "/internal/modules/catalog/adapters/http", "dependency direction"},
		{"cross domain", "internal/modules/work/application/create.go", mod + "/internal/modules/iam/domain", "other module api"},
		{"cross adapter", "internal/modules/work/adapters/http/handle.go", mod + "/internal/modules/iam/adapters/http", "other module api"},
		{"cross application", "internal/modules/work/application/create.go", mod + "/internal/modules/notifications/application", "other module api"},
		{"cross root", "internal/modules/work/wire.go", mod + "/internal/modules/iam", "other module api"},
		{"cross apis plural", "internal/modules/work/application/create.go", mod + "/internal/modules/iam/apis", "other module api"},
		{"forms package", "internal/modules/iam/application/create.go", mod + "/internal/modules/forms", "modules/forms and modules/workflow are forbidden"},
		{"forms api", "internal/modules/documents/adapters/s3blob/client.go", mod + "/core/internal/modules/forms/api", "modules/forms and modules/workflow are forbidden"},
		{"ai module", "internal/modules/iam/application/create.go", mod + "/internal/modules/ai", "modules/ai and modules/mcp are forbidden"},
		{"mcp service", "internal/modules/iam/application/create.go", mod + "/services/mcp/internal/config", "services/ai and services/mcp are forbidden"},
		{"workflow package", "internal/modules/work/application/create.go", mod + "/internal/modules/workflow/api", "modules/forms and modules/workflow are forbidden"},
		{"workflow other module", "cmd/setup/main.go", "github.com/example/kit/internal/modules/workflow", "modules/forms and modules/workflow are forbidden"},
		{"blob minio", "internal/modules/documents/adapters/s3blob/client.go", "github.com/minio/minio-go/v7", "MinIO SDK"},
		{"minio admin", "internal/platform/storage/client.go", "github.com/minio/madmin-go/v3", "MinIO SDK"},
		{"minio server", "internal/platform/storage/client.go", "github.com/minio/minio/pkg/madmin", "MinIO SDK"},
		{"blob rustfs", "internal/modules/documents/adapters/s3blob/client.go", "github.com/rustfs/client-go", "RustFS SDK"},
		{"rustfs sdk", "internal/platform/storage/client.go", "github.com/example/rustfs-sdk/s3", "RustFS SDK"},
		{"nonblob aws", "internal/modules/iam/application/create.go", "github.com/aws/aws-sdk-go-v2/service/s3", "limited to blob code"},
		{"setup aws", "cmd/setup/main.go", "github.com/aws/aws-sdk-go-v2/config", "limited to blob code"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeFixture(t, root, test.source, goImports(test.imported))
			violations := inspectFixture(t, root)
			if !containsViolation(violations, test.want) {
				t.Fatalf("missing %q in %v", test.want, violations)
			}
		})
	}
}

func TestInspectRejectsRemovedModuleSources(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "internal/modules/forms/domain/record.go", "package domain\n")
	writeFixture(t, root, "core/internal/modules/workflow/api/run.go", "package api\n")
	writeFixture(t, root, "internal/modules/ai/domain/prompt.go", "package domain\n")
	writeFixture(t, root, "services/mcp/cmd/mcp/main.go", "package main\n")
	violations := inspectFixture(t, root)
	for _, want := range []string{
		"internal/modules/forms/domain/record.go: modules/forms and modules/workflow are forbidden",
		"core/internal/modules/workflow/api/run.go: modules/forms and modules/workflow are forbidden",
		"internal/modules/ai/domain/prompt.go: modules/ai and modules/mcp are forbidden",
		"services/mcp/cmd/mcp/main.go: services/ai and services/mcp are forbidden",
	} {
		if !containsViolation(violations, want) {
			t.Fatalf("missing %q in %v", want, violations)
		}
	}
}

func TestInspectRejectsMalformedGo(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "internal/platform/http/edge.go", "package http\nimport (\n")
	violations := inspectFixture(t, root)
	if !containsViolation(violations, "cannot parse imports") {
		t.Fatalf("expected parse violation, got %v", violations)
	}
}
