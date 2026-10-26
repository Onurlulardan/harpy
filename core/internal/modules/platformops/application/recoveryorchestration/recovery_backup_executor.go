package recoveryorchestration

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	documentsapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api"
	iamapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api"
	authorityapi "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

const databaseBackupIdlePoll = 2 * time.Second

type databaseBackupExecutor struct {
	execution     RecoveryExecution
	snapshot      RecoveryTransition
	copy          RecoveryTransition
	verify        RecoveryTransition
	beginSnapshot func(context.Context) (RecoverySnapshot, error)
	newAttemptID  func(time.Time) (string, error)
	documents     documentsapi.Recovery
	iam           iamapi.RecoveryEvidence
	authority     authorityapi.RecoveryEvidence
	target        tenantapi.RecoveryTarget
	storage       BackupStorage
	archive       DatabaseArchive
	trust         recoverysigningv1.Bundle
	signer        ed25519.PrivateKey
	dsn           string
}

func (executor *databaseBackupExecutor) Run(
	ctx context.Context,
	admit func() bool,
	draining func() bool,
) error {
	if executor == nil || admit == nil || draining == nil {
		return fmt.Errorf("database backup executor is unavailable")
	}
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
		}
		if draining() {
			return nil
		}
		if !admit() {
			timer.Reset(databaseBackupIdlePoll)
			continue
		}
		processed, err := executor.processOne(ctx)
		if err != nil {
			return err
		}
		if draining() {
			return nil
		}
		if processed {
			timer.Reset(0)
		} else {
			timer.Reset(databaseBackupIdlePoll)
		}
	}
}

func (executor *databaseBackupExecutor) Close() {
	if executor != nil {
		clear(executor.signer)
		executor.signer = nil
	}
}

func (executor *databaseBackupExecutor) processOne(ctx context.Context) (bool, error) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	attemptID, err := executor.newAttemptID(now)
	if err != nil {
		return false, err
	}
	leaseToken, err := executor.newAttemptID(now)
	if err != nil {
		return false, err
	}
	claim, found, err := executor.execution.ClaimNext(
		ctx, attemptID, leaseToken,
	)
	if err != nil || !found {
		return found, err
	}
	identity := ports.RecoveryExecutionIdentity{
		RecoverySetID: claim.RecoverySetID,
		AttemptID:     claim.AttemptID, LeaseToken: claim.LeaseToken,
	}
	err = executor.withLease(ctx, identity, func(work context.Context) error {
		return executor.execute(work, claim, identity)
	})
	if err != nil && ctx.Err() == nil {
		_ = executor.execution.RejectFailure(ctx, ports.RecoveryExecutionFailure{
			RecoveryExecutionIdentity: identity,
			Code:                      "backup_execution_failed",
		})
	}
	return true, err
}
