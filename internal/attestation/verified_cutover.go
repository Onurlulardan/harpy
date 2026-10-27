package attestation

import (
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

type VerifiedCutoverCredentialLease struct {
	lease          cutoverv1.CredentialLease
	manifestSHA256 string
	canonical      []byte
	expiresAt      time.Time
	verified       bool
}

func (value VerifiedCutoverCredentialLease) LeaseID() string        { return value.lease.LeaseID }
func (value VerifiedCutoverCredentialLease) OperationID() string    { return value.lease.OperationID }
func (value VerifiedCutoverCredentialLease) RunID() string          { return value.lease.RunID }
func (value VerifiedCutoverCredentialLease) ManifestSHA256() string { return value.manifestSHA256 }
func (value VerifiedCutoverCredentialLease) PlanSHA256() string     { return value.lease.PlanSHA256 }
func (value VerifiedCutoverCredentialLease) TransformManifestSHA256() string {
	return value.lease.TransformManifestSHA256
}
func (value VerifiedCutoverCredentialLease) RelationCount() int {
	return len(value.lease.Relations)
}
func (value VerifiedCutoverCredentialLease) CanonicalDocument() []byte {
	return append([]byte(nil), value.canonical...)
}
func (value VerifiedCutoverCredentialLease) ExpiresAt() time.Time  { return value.expiresAt }
func (value VerifiedCutoverCredentialLease) EnvironmentID() string { return value.lease.EnvironmentID }
func (value VerifiedCutoverCredentialLease) ReleaseID() string     { return value.lease.ReleaseID }
func (value VerifiedCutoverCredentialLease) ScopeKind() string     { return value.lease.ScopeKind }
func (value VerifiedCutoverCredentialLease) ScopeID() string       { return value.lease.ScopeID }
func (value VerifiedCutoverCredentialLease) Source() cutoverv1.DatabaseIdentity {
	return value.lease.Source
}
func (value VerifiedCutoverCredentialLease) Target() cutoverv1.DatabaseIdentity {
	return value.lease.Target
}
func (value VerifiedCutoverCredentialLease) SourceConnectionLimit() int {
	return value.lease.SourceConnectionLimit
}
func (value VerifiedCutoverCredentialLease) TargetConnectionLimit() int {
	return value.lease.TargetConnectionLimit
}
func (value VerifiedCutoverCredentialLease) Relation(ordinal int) (cutoverv1.RelationMapping, bool) {
	if ordinal < 1 || ordinal > len(value.lease.Relations) {
		return cutoverv1.RelationMapping{}, false
	}
	relation := value.lease.Relations[ordinal-1]
	relation.Columns = append([]cutoverv1.ColumnMapping(nil), relation.Columns...)
	relation.ExpectedCheckpointKeys = append([]string(nil), relation.ExpectedCheckpointKeys...)
	return relation, relation.Ordinal == ordinal
}

type VerifiedCutoverSessionSeal struct {
	receipt       cutoverv1.SessionSealReceipt
	receiptSHA256 string
	canonical     []byte
	verified      bool
}

func (value VerifiedCutoverSessionSeal) ReceiptID() string {
	return value.receipt.Payload.ReceiptID
}
func (value VerifiedCutoverSessionSeal) ReceiptSHA256() string { return value.receiptSHA256 }
func (value VerifiedCutoverSessionSeal) CanonicalDocument() []byte {
	return append([]byte(nil), value.canonical...)
}
func (value VerifiedCutoverSessionSeal) SourceSessions() []cutoverv1.CutoverSession {
	return append([]cutoverv1.CutoverSession(nil), value.receipt.Payload.SourceSessions...)
}
func (value VerifiedCutoverSessionSeal) TargetSessions() []cutoverv1.CutoverSession {
	return append([]cutoverv1.CutoverSession(nil), value.receipt.Payload.TargetSessions...)
}

type VerifiedCutoverCredentialCleanup struct {
	receipt       cutoverv1.CredentialCleanupReceipt
	receiptSHA256 string
	canonical     []byte
	verified      bool
}

func (value VerifiedCutoverCredentialCleanup) ReceiptID() string {
	return value.receipt.Payload.ReceiptID
}
func (value VerifiedCutoverCredentialCleanup) ReceiptSHA256() string {
	return value.receiptSHA256
}
func (value VerifiedCutoverCredentialCleanup) CanonicalDocument() []byte {
	return append([]byte(nil), value.canonical...)
}
func (value VerifiedCutoverCredentialCleanup) IssuedAt() string {
	return value.receipt.Payload.IssuedAt
}

type VerifiedCutoverChangeBatch struct {
	batch       cutoverv1.ChangeBatch
	payloadJSON []byte
	verified    bool
}

func (value VerifiedCutoverChangeBatch) Payload() cutoverv1.ChangeBatchPayload {
	return value.batch.Payload
}
func (value VerifiedCutoverChangeBatch) PayloadJSON() []byte {
	return append([]byte(nil), value.payloadJSON...)
}
func (value VerifiedCutoverChangeBatch) SHA256() string { return value.batch.BatchSHA256 }

type VerifiedCutoverFinalizerLease struct {
	lease          cutoverv1.FinalizerLease
	manifestSHA256 string
	canonical      []byte
	expiresAt      time.Time
	verified       bool
}

func (value VerifiedCutoverFinalizerLease) LeaseID() string        { return value.lease.LeaseID }
func (value VerifiedCutoverFinalizerLease) OperationID() string    { return value.lease.OperationID }
func (value VerifiedCutoverFinalizerLease) RunID() string          { return value.lease.RunID }
func (value VerifiedCutoverFinalizerLease) ManifestSHA256() string { return value.manifestSHA256 }
func (value VerifiedCutoverFinalizerLease) CanonicalDocument() []byte {
	return append([]byte(nil), value.canonical...)
}
func (value VerifiedCutoverFinalizerLease) ExpiresAt() time.Time { return value.expiresAt }
func (value VerifiedCutoverFinalizerLease) Target() cutoverv1.DatabaseIdentity {
	return value.lease.Target
}
func (value VerifiedCutoverFinalizerLease) TerminalState() string { return value.lease.TerminalState }
func (value VerifiedCutoverFinalizerLease) TerminalEvidenceSHA256() string {
	return value.lease.TerminalEvidenceSHA256
}
func (value VerifiedCutoverFinalizerLease) ProcedureOID() uint32 {
	return value.lease.FinalizationProcedureOID
}
func (value VerifiedCutoverFinalizerLease) ProcedureSignature() string {
	return value.lease.FinalizationProcedureSignature
}
