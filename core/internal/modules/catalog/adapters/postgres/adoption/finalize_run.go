package adoption

import (
	"context"
	_ "embed"
	"time"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/adoption"
	"github.com/google/uuid"
)

//go:embed sql/tenant/record_run.sql
var recordRunSQL string

func (storage *store) FinalizeRun(ctx context.Context, value adoptionapp.Run) (adoptionapp.Run, error) {
	if !validRun(value) {
		return adoptionapp.Run{}, adoptionapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return adoptionapp.Run{}, unavailable()
	}
	var blocker any
	if value.BlockerCode != "" {
		blocker = value.BlockerCode
	}
	return scanRun(storage.queryRow(ctx, recordRunSQL, value.ID, value.SourceManifestSHA256,
		value.TransformSHA256, value.SourceCount, value.TargetCount, value.SourceRootSHA256,
		value.TargetRootSHA256, value.State, blocker, value.CutoverGeneration,
		value.CreatedAt, value.VerifiedAt))
}

func scanRun(source row) (adoptionapp.Run, error) {
	var value adoptionapp.Run
	var blocker *string
	var verified *time.Time
	if err := source.Scan(&value.ID, &value.SourceManifestSHA256, &value.TransformSHA256,
		&value.SourceCount, &value.TargetCount, &value.SourceRootSHA256,
		&value.TargetRootSHA256, &value.State, &blocker, &value.CutoverGeneration,
		&value.CreatedAt, &verified); err != nil {
		return adoptionapp.Run{}, err
	}
	if blocker != nil {
		value.BlockerCode = *blocker
	}
	value.CreatedAt = value.CreatedAt.UTC()
	if verified != nil {
		normalized := verified.UTC()
		verified = &normalized
	}
	value.VerifiedAt = verified
	if !validRun(value) {
		return adoptionapp.Run{}, unavailable()
	}
	return value, nil
}

func validRun(value adoptionapp.Run) bool {
	stateValid := value.State == "VERIFIED" && value.BlockerCode == "" &&
		value.VerifiedAt != nil || value.State == "BLOCKED" && value.BlockerCode != "" &&
		len(value.BlockerCode) <= 2048 && value.VerifiedAt == nil
	return value.ID != uuid.Nil && value.ID.Version() == 7 && value.SourceCount >= 0 &&
		value.TargetCount >= 0 && digestPattern.MatchString(value.SourceManifestSHA256) &&
		digestPattern.MatchString(value.TransformSHA256) &&
		digestPattern.MatchString(value.SourceRootSHA256) &&
		digestPattern.MatchString(value.TargetRootSHA256) && value.CutoverGeneration > 0 &&
		canonicalTime(value.CreatedAt) && (value.VerifiedAt == nil || canonicalTime(*value.VerifiedAt)) &&
		stateValid
}
