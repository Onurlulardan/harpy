package recoveryorchestration

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"io"
	"time"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	documentsapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api"
	iamapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api"
	authorityapi "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type RecoveryObjectReader interface {
	DownloadVerified(
		context.Context, ports.ObjectIdentity, io.Writer,
	) error
	OpenExact(context.Context, string, string) (io.ReadCloser, error)
}

type RecoveryObjectWriter interface {
	PutVerified(
		context.Context, ports.WriteRequest,
	) (ports.WriteEvidence, error)
}

type DatabaseArchive interface {
	Create(
		context.Context, string, string, io.Writer,
		backupcrypto.Context, backupcrypto.SealKeys,
	) (backupcrypto.Result, error)
	Restore(
		context.Context, string, io.Reader,
		backupcrypto.Context, backupcrypto.OpenKeys,
	) (int64, error)
	RestoreBound(
		context.Context, string, io.Reader,
		backupcrypto.BoundContext, backupcrypto.OpenKeyResolver,
	) (int64, backupcrypto.Context, error)
}

type RecoverySnapshot interface {
	Identifier() string
	Close() error
}

type RecoveryExecution interface {
	ClaimNext(
		context.Context, string, string,
	) (ports.RecoveryExecutionClaim, bool, error)
	UpdateLease(context.Context, ports.RecoveryExecutionIdentity) error
	AttestDocuments(context.Context, ports.DocumentsHoldAttestation) error
	PutComponents(context.Context, ports.RecoveryComponentBatch) error
	Complete(context.Context, ports.RecoveryCompletionRecord) error
	RejectFailure(context.Context, ports.RecoveryExecutionFailure) error
}

type RecoveryTransition interface {
	Start(context.Context, ports.RecoveryExecutionIdentity) error
}

type BackupStorage interface {
	Prefix(string, string) (string, error)
	RetentionDays() int
	RecoveryKeyProvider() (string, string)
	RecoveryReader() RecoveryObjectReader
	DocumentsReader() RecoveryObjectReader
	Writer(time.Time) RecoveryObjectWriter
}

type RestoreStorage interface {
	Prefix(string, string) (string, error)
	RecoveryKeyProvider() (string, string)
	RecoveryReader() RecoveryObjectReader
	RehydrateWriter(string) RecoveryObjectWriter
	ManifestVersions(context.Context, string, string) ([][]byte, error)
}

type RestoreBranch string

const (
	RestoreExistingReadable RestoreBranch = "existing_readable"
	RestoreRehydrate        RestoreBranch = "empty_or_unreadable"
)

type RestoreArtifacts struct {
	Branch                  RestoreBranch
	RecoveryStorageManifest string
	RecoveryTrustBundle     string
	OrchestratorTrustBundle string
	ProgressPath            string
	DrainAckPath            string
	SwitchAckPath           string
	ReceiptPath             string
	RestorePlan             string
	DRKEKKeyring            string
	ReceiptSigningKey       string
	RehydrateRequest        string
	RecoveryKit             string
}

type RestoreExecution interface {
	SeedRehydrate(context.Context, ports.RehydrateControlSeed) error
	Claim(
		context.Context, ports.RestoreExecutionIdentity, string,
	) (ports.RestoreExecutionClaim, bool, error)
	UpdateLease(context.Context, ports.RestoreExecutionIdentity) error
	StartDrain(context.Context, ports.RestoreExecutionIdentity) error
	ExecuteApply(context.Context, ports.RestoreApplyEvidence) error
	ValidateRestore(context.Context, ports.RestoreVerification) error
	MarkSwitched(context.Context, ports.RestoreSwitchEvidence) error
	Complete(context.Context, ports.RestoreCompletion) error
	RejectFailure(context.Context, ports.RestoreFailure) error
}

type RestoreScratch interface {
	Close()
	Documents() documentsapi.Recovery
	ImportReceipt(context.Context, ports.RestoreReceiptImport) error
}

type RestoreDependencies struct {
	Execution   RestoreExecution
	Storage     RestoreStorage
	Archive     DatabaseArchive
	OpenScratch func(
		context.Context, string, string, [32]byte, string,
	) (RestoreScratch, error)
	NewAttemptID  func(time.Time) (string, error)
	NewDatabaseID func() (string, error)
	LocatorSHA256 func(string) (string, error)
	Artifacts     RestoreArtifacts
	Environment   string
	ControlDSN    string
	ScratchDSN    string
}

type BackupDependencies struct {
	Execution         RecoveryExecution
	Snapshot          RecoveryTransition
	Copy              RecoveryTransition
	Verify            RecoveryTransition
	BeginSnapshot     func(context.Context) (RecoverySnapshot, error)
	NewAttemptID      func(time.Time) (string, error)
	Documents         documentsapi.Recovery
	IAMEvidence       iamapi.RecoveryEvidence
	AuthorityEvidence authorityapi.RecoveryEvidence
	Target            tenantapi.RecoveryTarget
	Storage           BackupStorage
	Archive           DatabaseArchive
	Trust             recoverysigningv1.Bundle
	Signer            ed25519.PrivateKey
	DSN               string
}

type BackupExecutor interface {
	Run(context.Context, func() bool, func() bool) error
	Close()
}

func NewBackupExecutor(dependencies BackupDependencies) (BackupExecutor, error) {
	if dependencies.Execution == nil || dependencies.Snapshot == nil ||
		dependencies.Copy == nil || dependencies.Verify == nil ||
		dependencies.BeginSnapshot == nil || dependencies.NewAttemptID == nil ||
		dependencies.Documents == nil || dependencies.IAMEvidence == nil ||
		dependencies.AuthorityEvidence == nil || dependencies.Storage == nil ||
		dependencies.Archive == nil || dependencies.Target.EnvironmentID() == "" ||
		dependencies.Trust.SHA256() == "" || len(dependencies.Signer) == 0 ||
		dependencies.DSN == "" {
		return nil, fmt.Errorf("database backup dependencies are incomplete")
	}
	return &databaseBackupExecutor{
		execution: dependencies.Execution,
		snapshot:  dependencies.Snapshot, copy: dependencies.Copy,
		verify: dependencies.Verify, beginSnapshot: dependencies.BeginSnapshot,
		newAttemptID: dependencies.NewAttemptID,
		documents:    dependencies.Documents, iam: dependencies.IAMEvidence,
		authority: dependencies.AuthorityEvidence, target: dependencies.Target,
		storage: dependencies.Storage, archive: dependencies.Archive,
		trust: dependencies.Trust, signer: dependencies.Signer,
		dsn: dependencies.DSN,
	}, nil
}
