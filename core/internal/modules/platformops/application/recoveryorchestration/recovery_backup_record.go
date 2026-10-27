package recoveryorchestration

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseBackupExecutor) recordComponentBatch(
	ctx context.Context,
	identity ports.RecoveryExecutionIdentity,
	components map[string]databaseBackupComponent,
) error {
	inventory := make(map[string]recoveryset.ComponentIdentity, len(components))
	for kind, component := range components {
		if kind != component.kind {
			return fmt.Errorf("recovery component kind differs")
		}
		inventory[kind] = component.manifestIdentity()
	}
	document, err := json.Marshal(inventory)
	if err != nil {
		return err
	}
	return executor.execution.PutComponents(
		ctx, ports.RecoveryComponentBatch{
			RecoveryExecutionIdentity: identity,
			Components:                document,
		},
	)
}

func recoveryComponents(
	components map[string]databaseBackupComponent,
) (recoveryset.RecoveryComponents, error) {
	get := func(kind string) (recoveryset.ComponentIdentity, bool) {
		component, found := components[kind]
		return component.manifestIdentity(), found && component.kind == kind
	}
	database, databaseOK := get("database_archive")
	documents, documentsOK := get("documents_inventory")
	custody, custodyOK := get("key_custody")
	owner, ownerOK := get("owner_anchor")
	authority, authorityOK := get("protected_authority")
	if !databaseOK || !documentsOK || !custodyOK || !ownerOK || !authorityOK {
		return recoveryset.RecoveryComponents{},
			fmt.Errorf("recovery component set is incomplete")
	}
	return recoveryset.RecoveryComponents{
		DatabaseArchive: database, DocumentsInventory: documents,
		KeyCustody: custody, OwnerAnchor: owner,
		ProtectedAuthority: authority,
	}, nil
}
