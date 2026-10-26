package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/jackc/pgx/v5"
)

type FinalizationDatabase struct {
	connection     *pgx.Conn
	cleanupTimeout time.Duration
	ready          bool
}

type FinalizationInput struct {
	AdoptionLease    attestation.VerifiedCutoverCredentialLease
	SessionSeal      attestation.VerifiedCutoverSessionSeal
	Cleanup          attestation.VerifiedCutoverCredentialCleanup
	Authority        attestation.VerifiedAuthorityAdoptionReceipt
	FinalizerLease   attestation.VerifiedCutoverFinalizerLease
	TerminalEvidence attestation.VerifiedCutoverTerminalEvidence
}

type FinalizationResult struct {
	TerminalState          string
	TerminalEvidenceSHA256 string
}

func OpenFinalizationDatabase(
	ctx context.Context,
	settings config.PrivilegedPostgres,
	lease attestation.VerifiedCutoverFinalizerLease,
) (*FinalizationDatabase, error) {
	if lease.OperationID() == "" || lease.RunID() == "" || lease.ProcedureOID() == 0 {
		return nil, fmt.Errorf("cutover finalizer lease is invalid")
	}
	applicationName := cutoverFinalizerApplicationName(lease.RunID())
	connection, err := openConnection(ctx, settings, applicationName)
	if err != nil {
		return nil, err
	}
	database := &FinalizationDatabase{
		connection: connection, cleanupTimeout: settings.CleanupTimeout(),
	}
	if err := database.verifyAuthority(ctx, lease, applicationName); err != nil {
		_ = database.Close()
		return nil, err
	}
	database.ready = true
	return database, nil
}

func (database *FinalizationDatabase) verifyAuthority(
	ctx context.Context,
	lease attestation.VerifiedCutoverFinalizerLease,
	applicationName string,
) error {
	observation, err := observeSession(ctx, database.connection)
	if err != nil || observation.validateDatabase(lease.Target()) != nil ||
		observation.sessionUser != "harpy_cutover_finalizer" ||
		observation.applicationName != applicationName {
		return fmt.Errorf("cutover finalizer database authority differs")
	}
	var procedureOID int64
	var procedureSignature string
	err = database.connection.QueryRow(ctx, observeFinalizerProcedureSQL).Scan(
		&procedureOID, &procedureSignature,
	)
	if err != nil || procedureOID != int64(lease.ProcedureOID()) ||
		procedureSignature != lease.ProcedureSignature() {
		return fmt.Errorf("cutover finalizer procedure binding differs")
	}
	return nil
}

func (database *FinalizationDatabase) Finalize(
	ctx context.Context,
	input FinalizationInput,
) (FinalizationResult, error) {
	lease := input.FinalizerLease
	if database == nil || !database.ready || input.AdoptionLease.OperationID() != lease.OperationID() ||
		input.AdoptionLease.RunID() != lease.RunID() {
		return FinalizationResult{}, fmt.Errorf("cutover finalization authority differs")
	}
	finalizerDocument, adoptionDocument, sealDocument, cleanupDocument, authorityDocument,
		terminalDocument, err :=
		finalizationDocuments(input)
	if err != nil {
		return FinalizationResult{}, err
	}
	targetInput := targetFinalizationInput{
		operationID: lease.OperationID(), runID: lease.RunID(), leaseID: lease.LeaseID(),
		manifestSHA256: lease.ManifestSHA256(), finalizerDocument: finalizerDocument,
		adoptionSHA256: input.AdoptionLease.ManifestSHA256(), adoptionDocument: adoptionDocument,
		sealSHA256: input.SessionSeal.ReceiptSHA256(), sealDocument: sealDocument,
		cleanupSHA256: input.Cleanup.ReceiptSHA256(), cleanupDocument: cleanupDocument,
		authoritySHA256: input.Authority.ReceiptSHA256(), authorityDocument: authorityDocument,
		terminalState: lease.TerminalState(), terminalEvidenceSHA256: lease.TerminalEvidenceSHA256(),
		terminalEvidenceDocument: terminalDocument,
	}
	return finalizeTargetTransaction(ctx, database.connection, database.cleanupTimeout, targetInput)
}

func finalizationDocuments(
	input FinalizationInput,
) ([]byte, []byte, []byte, []byte, []byte, []byte, error) {
	finalizerDocument := input.FinalizerLease.CanonicalDocument()
	adoptionDocument := input.AdoptionLease.CanonicalDocument()
	sealDocument := input.SessionSeal.CanonicalDocument()
	cleanupDocument := input.Cleanup.CanonicalDocument()
	authorityDocument := input.Authority.CanonicalDocument()
	terminalDocument := input.TerminalEvidence.CanonicalDocument()
	bindings := []struct {
		document []byte
		digest   string
	}{
		{finalizerDocument, input.FinalizerLease.ManifestSHA256()},
		{adoptionDocument, input.AdoptionLease.ManifestSHA256()},
		{sealDocument, input.SessionSeal.ReceiptSHA256()},
		{cleanupDocument, input.Cleanup.ReceiptSHA256()},
		{authorityDocument, input.Authority.ReceiptSHA256()},
	}
	for _, binding := range bindings {
		digest, err := attestation.DigestHex(binding.document)
		if err != nil || digest != binding.digest {
			return nil, nil, nil, nil, nil, nil,
				fmt.Errorf("cutover verified document binding differs")
		}
	}
	if input.TerminalEvidence.SHA256() != input.FinalizerLease.TerminalEvidenceSHA256() ||
		input.TerminalEvidence.Value().TerminalState != input.FinalizerLease.TerminalState() ||
		len(terminalDocument) == 0 {
		return nil, nil, nil, nil, nil, nil,
			fmt.Errorf("cutover terminal evidence binding differs")
	}
	return finalizerDocument, adoptionDocument, sealDocument, cleanupDocument,
		authorityDocument, terminalDocument, nil
}

func (database *FinalizationDatabase) Close() error {
	if database == nil || database.connection == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), database.cleanupTimeout)
	defer cancel()
	err := closeConnections(ctx, []*pgx.Conn{database.connection})
	database.connection = nil
	return err
}
