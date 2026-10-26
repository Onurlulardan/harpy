//go:build integration

package main

import (
	"context"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type workflowFinalizationFixture struct {
	ids              workflowCutoverIDs
	state            string
	prerequisites    finalizationPrerequisitePaths
	clusterDSNPath   string
	finalizerDSNPath string
	leasePath        string
	signaturePath    string
	signerPath       string
}

func prepareWorkflowFinalizationFixture(
	t *testing.T, database workflowDatabase, authority workflowAuthorityFixture,
	directory, state string, now time.Time,
) workflowFinalizationFixture {
	t.Helper()
	ids := workflowIDsForState(state)
	lease := workflowCredentialLease(t, database.target, ids, now)
	leasePath, leaseDocument := writeWorkflowArtifact(
		t, directory, state+"-credential-lease.json", lease,
	)
	leaseSignature := signWorkflowDocument(t, authority.signer, leaseDocument,
		now.Add(-2*time.Minute), "cutover-credential-lease-signature.v1")
	leaseSignaturePath := writeWorkflowBytes(
		t, directory, state+"-credential-lease.sig", leaseSignature,
	)
	leaseSHA, err := attestation.DigestHex(leaseDocument)
	if err != nil {
		t.Fatal(err)
	}
	sealDocument := workflowSessionSeal(
		t, authority.signer, lease, leaseSHA, ids, now,
	)
	sealPath := writeWorkflowBytes(t, directory, state+"-seal.json", sealDocument)
	sealSHA, err := attestation.DigestHex(sealDocument)
	if err != nil {
		t.Fatal(err)
	}
	cleanupDocument := workflowCleanup(
		t, authority.signer, lease, leaseSHA, sealSHA, ids, now,
	)
	cleanupPath := writeWorkflowBytes(
		t, directory, state+"-cleanup.json", cleanupDocument,
	)
	terminal := workflowTerminalEvidence(
		state, ids, leaseSHA, lease.PlanSHA256, lease.TransformManifestSHA256,
	)
	if err := cutoverv1.ValidateTerminalEvidence(terminal); err != nil {
		t.Fatalf("validate %s terminal evidence: %v", state, err)
	}
	terminalPath, _ := writeWorkflowArtifact(
		t, directory, state+"-terminal.json", terminal,
	)
	clusterDSNPath := writeWorkflowBytes(t, directory, state+"-cluster.dsn",
		[]byte(workflowRoleDSN(t, database.dsn, "harpy_cluster_admin")))
	finalizerDSNPath := writeWorkflowBytes(t, directory, state+"-finalizer.dsn",
		[]byte(workflowRoleDSN(t, database.dsn, "harpy_cutover_finalizer")))
	return workflowFinalizationFixture{ids: ids, state: state,
		clusterDSNPath: clusterDSNPath, finalizerDSNPath: finalizerDSNPath,
		leasePath:     directory + "/" + state + "-finalizer-lease.json",
		signaturePath: directory + "/" + state + "-finalizer-lease.sig",
		signerPath:    authority.signer.signerPath,
		prerequisites: finalizationPrerequisitePaths{
			authorityArtifactPaths:    authority.paths,
			AuthorityReceipt:          authority.receiptPath,
			AuthorityReceiptSignature: authority.signaturePath,
			CredentialLease:           leasePath, CredentialLeaseSignature: leaseSignaturePath,
			SessionSealReceipt: sealPath, CredentialCleanupReceipt: cleanupPath,
			TerminalEvidence: terminalPath,
		},
	}
}

func (value workflowFinalizationFixture) issueInvocation() invocation {
	return invocation{ContractVersion: invocationContract, Action: "issue_finalizer",
		Environment: "test", DeadlineSeconds: 120,
		IssueFinalizer: &finalizerIssueInvocation{
			ClusterAdminDSNFile:           value.clusterDSNPath,
			FinalizerDSNFile:              value.finalizerDSNPath,
			finalizationPrerequisitePaths: value.prerequisites,
			SigningKeyFile:                value.signerPath, FinalizerLeaseID: value.ids.finalizer,
			FinalizerLeasePath:          value.leasePath,
			FinalizerLeaseSignaturePath: value.signaturePath,
		}}
}

func (value workflowFinalizationFixture) authorizeInvocation() invocation {
	return invocation{ContractVersion: invocationContract, Action: "authorize_finalization",
		Environment: "test", DeadlineSeconds: 120,
		AuthorizeFinalization: &authorizationInvocation{
			ClusterAdminDSNFile:           value.clusterDSNPath,
			finalizationPrerequisitePaths: value.prerequisites,
			FinalizerLease:                value.leasePath,
			FinalizerLeaseSignature:       value.signaturePath,
		}}
}

func adoptWorkflowAuthority(
	t *testing.T, ctx context.Context, database workflowDatabase,
	directory string, now time.Time,
) workflowAuthorityFixture {
	t.Helper()
	authority := prepareWorkflowAuthorityArtifacts(t, ctx, database, directory, now)
	result, err := executeAuthorityAdoption(ctx, authority.adoptionInvocation())
	if err != nil || result.Replayed {
		t.Fatalf("adopt workflow authority: result=%+v err=%v", result, err)
	}
	return authority
}
