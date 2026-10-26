package recoverysigningv1

const (
	BundleSchema = "recovery-trust-bundle.v1"
	Algorithm    = "Ed25519"
	Generation   = int64(1)

	BundleFilename         = "recovery-trust-bundle.v1.json"
	ManifestSignerFilename = "recovery-manifest-signer.ed25519"
	ControlSignerFilename  = "recovery-control-signer.ed25519"
	ReceiptSignerFilename  = "restore-receipt-signer.ed25519"
	MaximumBundleBytes     = 32 * 1024
)

type Purpose string

const (
	PurposeRecoveryManifest Purpose = "recovery_manifest_signer"
	PurposeRecoveryControl  Purpose = "recovery_control_signer"
	PurposeRestoreReceipt   Purpose = "restore_receipt_signer"
)

var orderedPurposes = []Purpose{
	PurposeRecoveryManifest,
	PurposeRecoveryControl,
	PurposeRestoreReceipt,
}

func Filename(purpose Purpose) string {
	switch purpose {
	case PurposeRecoveryManifest:
		return ManifestSignerFilename
	case PurposeRecoveryControl:
		return ControlSignerFilename
	case PurposeRestoreReceipt:
		return ReceiptSignerFilename
	default:
		return ""
	}
}
