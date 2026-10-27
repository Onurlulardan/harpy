package main

import (
	"crypto/sha256"
	"fmt"

	corecutover "github.com/Onurlulardan/harpy/core/public/cutover"
)

func verifyFullRolloutArtifacts(
	paths []string, evidence corecutover.FullRolloutEvidence,
) error {
	expected := fullRolloutArtifactDigests(evidence)
	for _, path := range paths {
		document, err := readArtifact(path, maximumArtifactBytes)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(document)
		clear(document)
		if _, found := expected[digest]; !found {
			return fmt.Errorf("completion artifact is not bound to evidence")
		}
		delete(expected, digest)
	}
	if len(expected) != 0 {
		return fmt.Errorf("completion evidence artifact is missing")
	}
	return nil
}

func fullRolloutArtifactDigests(
	evidence corecutover.FullRolloutEvidence,
) map[[32]byte]struct{} {
	result := make(map[[32]byte]struct{})
	for _, value := range [][32]byte{
		evidence.ShadowReadSHA256, evidence.LogicalDeltaSHA256,
		evidence.ReconcileSHA256, evidence.ProductionSnapshotSHA256,
		evidence.V1RemovalSHA256, evidence.BackupRestoreSHA256,
		evidence.FrontendContractSHA256, evidence.MCPContractSHA256,
		evidence.RollbackWindowClosureSHA256,
	} {
		result[value] = struct{}{}
	}
	for _, step := range evidence.Steps {
		for _, value := range [][32]byte{step.AcceptanceSHA256,
			step.ReverseRehearsalSHA256, step.SourceSHA256, step.TargetSHA256} {
			result[value] = struct{}{}
		}
	}
	delete(result, [32]byte{})
	return result
}
