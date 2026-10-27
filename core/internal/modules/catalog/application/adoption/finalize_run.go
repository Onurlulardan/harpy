package adoption

import "context"

func FinalizeRun(ctx context.Context, store Store, run Run) (Run, error) {
	if ctx == nil || ctx.Err() != nil || store == nil {
		return Run{}, ErrUnavailable
	}
	stored, err := store.FinalizeRun(ctx, run)
	if err != nil {
		return Run{}, err
	}
	if !sameRun(stored, run) {
		return Run{}, ErrUnavailable
	}
	return stored, nil
}

func sameRun(left, right Run) bool {
	if left.ID != right.ID || left.SourceManifestSHA256 != right.SourceManifestSHA256 ||
		left.TransformSHA256 != right.TransformSHA256 || left.SourceCount != right.SourceCount ||
		left.TargetCount != right.TargetCount || left.SourceRootSHA256 != right.SourceRootSHA256 ||
		left.TargetRootSHA256 != right.TargetRootSHA256 || left.State != right.State ||
		left.BlockerCode != right.BlockerCode || left.CutoverGeneration != right.CutoverGeneration ||
		!left.CreatedAt.Equal(right.CreatedAt) || (left.VerifiedAt == nil) != (right.VerifiedAt == nil) {
		return false
	}
	return left.VerifiedAt == nil || left.VerifiedAt.Equal(*right.VerifiedAt)
}
