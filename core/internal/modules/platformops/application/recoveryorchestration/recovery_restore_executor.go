package recoveryorchestration

import (
	"context"
	"fmt"

	documentsapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type databaseRestoreExecutor struct {
	execution   RestoreExecution
	storage     RestoreStorage
	authority   *databaseRestoreAuthority
	keys        databaseRestoreKeys
	archive     DatabaseArchive
	artifacts   RestoreArtifacts
	controlDSN  string
	scratchDSN  string
	openScratch func(
		context.Context, string, string, [32]byte, string,
	) (RestoreScratch, error)
	newDatabaseID func() (string, error)
	locatorSHA256 func(string) (string, error)
}

type databaseRestoreEvidence struct {
	archiveSHA256      string
	newDatabaseID      string
	objectMerkle       string
	locatorRemap       string
	safety             string
	componentInventory []byte
	scratch            RestoreScratch
	documents          documentsapi.Recovery
}

func (evidence *databaseRestoreEvidence) close() {
	if evidence != nil && evidence.scratch != nil {
		evidence.scratch.Close()
		evidence.scratch = nil
	}
}

func (executor *databaseRestoreExecutor) execute(
	ctx context.Context,
	claim ports.RestoreExecutionClaim,
	identity ports.RestoreExecutionIdentity,
) error {
	manifest, err := executor.validateClaim(claim, identity)
	if err != nil {
		return terminalRestore(err)
	}
	if err = executor.verifySafetySet(ctx, claim, manifest); err != nil {
		return err
	}
	controlLocator, scratchLocator, err := restoreLocators(
		executor.controlDSN, executor.scratchDSN, executor.locatorSHA256,
	)
	if err != nil {
		return terminalRestore(err)
	}
	if claim.RestoreState == "plan_ready" {
		if err = executor.execution.StartDrain(ctx, identity); err != nil {
			return resumableRestore(err)
		}
		claim.RestoreState = "draining"
	}
	if claim.RestoreState == "draining" {
		if err = executor.enterRestoreApply(
			ctx, &claim, identity, controlLocator, scratchLocator,
		); err != nil {
			return err
		}
	}
	var evidence databaseRestoreEvidence
	if claim.RestoreState == "restoring" || claim.RestoreState == "verifying" {
		evidence, err = executor.prepareScratch(
			ctx, claim, identity, manifest,
			claim.RestoreState == "restoring",
		)
		if err != nil {
			return err
		}
		defer evidence.close()
	}
	if claim.RestoreState == "restoring" {
		if err = executor.execution.ValidateRestore(ctx, ports.RestoreVerification{
			RestoreExecutionIdentity: identity,
			ArchiveReadbackSHA256:    evidence.archiveSHA256,
			NewDatabaseSystemID:      evidence.newDatabaseID,
		}); err != nil {
			return resumableRestore(err)
		}
		claim.RestoreState = "verifying"
		claim.ArchiveReadbackSHA256 = evidence.archiveSHA256
		claim.NewDatabaseSystemID = evidence.newDatabaseID
	}
	if claim.RestoreState == "verifying" {
		if err = executor.enterRestoreSwitch(
			ctx, &claim, identity, evidence, controlLocator, scratchLocator,
		); err != nil {
			return err
		}
	}
	if claim.RestoreState != "switched" {
		return terminalRestore(fmt.Errorf("restore state is unsupported"))
	}
	return executor.completeRestore(
		ctx, claim, identity, manifest, controlLocator, scratchLocator,
	)
}
