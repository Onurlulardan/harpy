package demomaterialization

import (
	"bytes"
	"time"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
)

func snapshotMatches(
	snapshot Snapshot,
	request verifiedRequest,
	requestSHA256 [32]byte,
) bool {
	if snapshot.ReceiptID.Version() != 7 ||
		snapshot.PackageID != request.packageID ||
		snapshot.PackageSlug != request.request.PackageSlug ||
		snapshot.PackageVersion != request.request.PackageVersion ||
		snapshot.ReleaseID != request.releaseID ||
		snapshot.Component != request.request.Component ||
		!sameMaterialization(snapshot.Materialization, request.materialization) ||
		snapshot.IdempotencyKey != request.request.IdempotencyKey ||
		snapshot.RequestSHA256 != requestSHA256 ||
		(snapshot.State != contract.ProvisionAvailable &&
			snapshot.State != contract.ProvisionStaged) ||
		!canonicalTime(snapshot.CompletedAt) ||
		!canonicalTime(snapshot.CreatedAt) ||
		!canonicalTime(snapshot.UpdatedAt) ||
		snapshot.UpdatedAt.Before(snapshot.CreatedAt) ||
		snapshot.CompletedAt.Before(snapshot.CreatedAt) {
		return false
	}
	if snapshot.Ready {
		return snapshot.RetryAt.IsZero()
	}
	return canonicalTime(snapshot.RetryAt) &&
		snapshot.RetryAt.After(snapshot.CompletedAt)
}

func sameMaterialization(left, right Materialization) bool {
	return left.ModuleSlug == right.ModuleSlug &&
		left.SourceContract == right.SourceContract &&
		bytes.Equal(left.Document, right.Document) &&
		left.SemanticSHA256 == right.SemanticSHA256 &&
		left.FormCount == right.FormCount &&
		left.RowCount == right.RowCount
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%1000 == 0
}
