package adoption

import (
	"context"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
)

func (storage *Store) VerifyCopy(ctx context.Context, authority adoptionapp.Authority,
	checkpoint adoptionapp.Checkpoint, receipt adoptionapp.CopyReceipt,
) (adoptionapp.Checkpoint, error) {
	object := receipt.TargetObject
	if !storage.usable(ctx) || !storage.owns(authority) ||
		checkpoint.RunID != authority.RunID || !checkpoint.SourceVersionID.UUIDv7() ||
		!checkpoint.TargetVersionID.UUIDv7() || object.Key == "" || object.Version == "" ||
		object.Generation == 0 || object.PlaintextBytes != checkpoint.SourceBytes ||
		object.PlaintextSHA256 != checkpoint.SourceSHA256 ||
		object.CiphertextBytes <= object.PlaintextBytes ||
		object.CiphertextSHA256 == [32]byte{} || object.KeyID == "" ||
		len(object.WrappedDEK) < 115 || receipt.MetadataSHA256 == [32]byte{} ||
		receipt.AppliedLSN == "" || receipt.VerifiedAt.IsZero() {
		return adoptionapp.Checkpoint{}, adoptionapp.ErrInvalid
	}
	result, _, err := storage.scanCheckpoint(storage.queryRow(ctx, verifyCopySQL,
		authority.RunID.String(), int64(authority.CutoverGeneration),
		authority.ManifestSHA256[:], checkpoint.SourceVersionID.String(),
		checkpoint.SourceRevision, checkpoint.TargetVersionID.String(), object.Key,
		object.Version, int64(object.Generation), object.CiphertextBytes,
		object.PlaintextSHA256[:], object.CiphertextSHA256[:], object.KeyID,
		object.WrappedDEK, receipt.MetadataSHA256[:], receipt.AppliedLSN,
		receipt.VerifiedAt))
	return result, err
}
