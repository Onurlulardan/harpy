//go:build integration

package main

import (
	"context"
	"testing"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func workflowDependencies(
	t *testing.T, ctx context.Context, query catalogQuerier,
) []cutoverv1.AuthorityCatalogDependency {
	t.Helper()
	identities := make([]workflowIdentity, 0, len(workflowDependencyTables)+1)
	identities = append(identities, workflowIdentity{"schema", "platform", "platform", ""})
	for _, table := range workflowDependencyTables {
		identities = append(identities, workflowIdentity{"table", "platform", table, ""})
	}
	dependencies := make([]cutoverv1.AuthorityCatalogDependency, len(identities))
	for index, identity := range identities {
		state, err := readCatalogState(ctx, query, identity.kind, identity.schema,
			identity.name, identity.arguments)
		if err != nil {
			t.Fatal(err)
		}
		dependencies[index] = cutoverv1.AuthorityCatalogDependency{
			Kind: identity.kind, SchemaName: identity.schema, ObjectName: identity.name,
			IdentityArguments: identity.arguments, Owner: state.owner,
			DefinitionSHA256: state.definition, ACLSHA256: state.acl,
		}
	}
	return dependencies
}
