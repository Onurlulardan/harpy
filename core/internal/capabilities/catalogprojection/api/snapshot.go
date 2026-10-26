package catalogprojection

import (
	"bytes"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
)

func snapshotMatches(
	snapshot Snapshot,
	request verifiedRequest,
	requestSHA256 [32]byte,
) bool {
	return snapshot.ReceiptID.Version() == 7 &&
		snapshot.PackageID == request.packageID &&
		snapshot.PackageSlug == request.request.PackageSlug &&
		snapshot.PackageVersion == request.request.PackageVersion &&
		snapshot.ReleaseID == request.releaseID &&
		snapshot.Component == request.request.Component &&
		sameMaterialization(snapshot.Materialization, request.materialization) &&
		snapshot.IdempotencyKey == request.request.IdempotencyKey &&
		snapshot.RequestSHA256 == requestSHA256 &&
		(snapshot.State == contract.ProvisionAvailable ||
			snapshot.State == contract.ProvisionStaged) &&
		!snapshot.CreatedAt.IsZero() && !snapshot.UpdatedAt.IsZero() &&
		!snapshot.UpdatedAt.Before(snapshot.CreatedAt) &&
		(snapshot.ReadyAt.IsZero() ||
			!snapshot.ReadyAt.Before(snapshot.CreatedAt))
}

func sameMaterialization(left, right Materialization) bool {
	return left.Kind == right.Kind && left.ModuleSlug == right.ModuleSlug &&
		left.Identity == right.Identity &&
		bytes.Equal(left.Document, right.Document) &&
		left.SemanticSHA256 == right.SemanticSHA256 &&
		left.ItemCount == right.ItemCount &&
		left.AuxiliaryCount == right.AuxiliaryCount
}
