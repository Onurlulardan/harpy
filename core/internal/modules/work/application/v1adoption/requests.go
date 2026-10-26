package v1adoption

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type ApplyRequest struct {
	RunID          domain.ID
	Generation     int64
	ManifestSHA256 [sha256.Size]byte
}

func (request ApplyRequest) Valid() bool {
	return request.RunID.UUIDv7() && request.Generation > 0 &&
		request.ManifestSHA256 != [sha256.Size]byte{}
}

type ReconcileRequest = ApplyRequest

type ReverseRequest struct {
	RunID, ReconciliationID, ApprovedBy domain.ID
	SourceGeneration, TargetGeneration  int64
}

func (request ReverseRequest) Valid() bool {
	return request.RunID.Valid() && request.ReconciliationID.Valid() && request.ApprovedBy.Valid() &&
		request.SourceGeneration > 0 && request.TargetGeneration > 0 &&
		request.SourceGeneration != request.TargetGeneration
}

type RetireRequest struct{ RunID domain.ID }

func (request RetireRequest) Valid() bool { return request.RunID.Valid() }
