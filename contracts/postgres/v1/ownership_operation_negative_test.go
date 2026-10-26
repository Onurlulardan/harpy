package postgresv1

import (
	"strings"
	"testing"
)

func TestOwnershipManifestRejectsBrokenTreeContracts(t *testing.T) {
	inlineTests := []mutationCase{
		{"wrong executor", func(value map[string]any) {
			value["ownership_executor"] = "external_root_provisioner"
		}},
		{"both encodings", func(value map[string]any) {
			value["chunks"] = validOwnershipManifest("fresh", "chunked")["chunks"]
		}},
		{"unknown object field", func(value map[string]any) {
			value["objects"].([]any)[0].(map[string]any)["sql"] = "ALTER DATABASE"
		}},
		{"duplicate object", func(value map[string]any) {
			objects := value["objects"].([]any)
			value["objects"] = append(objects, objects[0])
		}},
	}
	assertMutationsInvalid(t, "ownership-manifest.schema.json", func() map[string]any {
		return validOwnershipManifest("fresh", "inline")
	}, inlineTests)

	chunkedTests := []mutationCase{
		{"oversize chunk", func(value map[string]any) {
			ownershipDescriptor(value)["length_bytes"] = 4194305
		}},
		{"traversal path", func(value map[string]any) {
			ownershipDescriptor(value)["path"] = "../chunk.json"
		}},
		{"uppercase hash", func(value map[string]any) {
			ownershipDescriptor(value)["sha256"] = strings.Repeat("A", 64)
		}},
	}
	assertMutationsInvalid(t, "ownership-manifest.schema.json", func() map[string]any {
		return validOwnershipManifest("v1_adopted", "chunked")
	}, chunkedTests)
}

func TestOwnershipChunkRejectsUnsafeContent(t *testing.T) {
	tests := []mutationCase{
		{"negative ordinal", func(value map[string]any) { value["ordinal"] = -1 }},
		{"too many chunks", func(value map[string]any) { value["count"] = 513 }},
		{"unknown object kind", func(value map[string]any) {
			ownershipChunkObject(value)["object"].(map[string]any)["kind"] = "raw_sql"
		}},
		{"null owner", func(value map[string]any) {
			ownershipChunkObject(value)["owner"] = nil
		}},
	}
	assertMutationsInvalid(t, "ownership-chunk.schema.json", func() map[string]any {
		return validOwnershipChunk("fresh")
	}, tests)
}

func TestBootstrapOperationRejectsStepDrift(t *testing.T) {
	tests := []mutationCase{
		{"wrong origin sequence", func(value map[string]any) {
			value["database_origin"] = "v1_adopted"
		}},
		{"reordered steps", func(value map[string]any) {
			steps := value["steps"].([]any)
			steps[0], steps[1] = steps[1], steps[0]
		}},
		{"arbitrary sql", func(value map[string]any) {
			value["steps"].([]any)[3].(map[string]any)["sql"] = "CREATE ROLE bypass"
		}},
		{"wrong transaction class", func(value map[string]any) {
			value["steps"].([]any)[2].(map[string]any)["transaction_class"] =
				"transactional_ddl"
		}},
		{"invalid input hash", func(value map[string]any) {
			value["input_hashes"].(map[string]any)["roles_manifest_sha256"] = digest("A")
		}},
	}
	assertMutationsInvalid(t, "bootstrap-operation.schema.json", func() map[string]any {
		return validBootstrapOperation("fresh")
	}, tests)
}

func TestExtensionOperationRejectsStepDrift(t *testing.T) {
	tests := []mutationCase{
		{"unknown step", func(value map[string]any) {
			value["steps"].([]any)[3].(map[string]any)["step_id"] = "run_sql"
		}},
		{"wrong transaction class", func(value map[string]any) {
			value["steps"].([]any)[4].(map[string]any)["transaction_class"] =
				"autocommit_ddl"
		}},
		{"missing ownership hash", func(value map[string]any) {
			delete(value["input_hashes"].(map[string]any), "ownership_manifest_sha256")
		}},
	}
	assertMutationsInvalid(t, "extension-operation.schema.json", validExtensionOperation, tests)
}

func TestCleanupOperationRejectsAuthorityDrift(t *testing.T) {
	tests := []mutationCase{
		{"role order", func(value map[string]any) {
			roles := value["privileged_roles"].([]any)
			roles[0], roles[1] = roles[1], roles[0]
		}},
		{"wrong step class", func(value map[string]any) {
			value["steps"].([]any)[1].(map[string]any)["transaction_class"] =
				"transactional_ddl"
		}},
		{"arbitrary sql", func(value map[string]any) {
			value["steps"].([]any)[1].(map[string]any)["sql"] = "ALTER ROLE"
		}},
	}
	assertMutationsInvalid(
		t, "credential-cleanup-operation.schema.json", validCleanupOperation, tests,
	)
}

func ownershipDescriptor(value map[string]any) map[string]any {
	return value["chunks"].([]any)[0].(map[string]any)
}

func ownershipChunkObject(value map[string]any) map[string]any {
	return value["objects"].([]any)[0].(map[string]any)
}
