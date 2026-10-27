package recoverysigning

import (
	"context"
	"os"
	"path/filepath"
	"time"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func generateAndPublish(
	ctx context.Context,
	target string,
	environmentID string,
	generatedAt time.Time,
	deps dependencies,
) (result Result, err error) {
	staging := target + ".staging"
	keepStaging := false
	defer func() {
		if err != nil && !keepStaging {
			_ = os.RemoveAll(staging)
		}
	}()
	if err = os.Mkdir(staging, 0o700); err != nil {
		return Result{}, ErrFenced
	}
	set, err := generate(ctx, environmentID, generatedAt, deps)
	if err != nil {
		return Result{}, err
	}
	for purpose, seed := range set.seeds {
		if err = writeSynced(filepath.Join(staging, recoveryv1.Filename(purpose)), seed, 0o600); err != nil {
			return Result{}, err
		}
	}
	if err = writeSynced(filepath.Join(staging, recoveryv1.BundleFilename),
		set.bundle.CanonicalBytes(), 0o644); err != nil {
		return Result{}, err
	}
	if err = syncDirectory(staging); err != nil {
		return Result{}, err
	}
	if deps.beforePublish != nil {
		if hookErr := deps.beforePublish(); hookErr != nil {
			keepStaging = true
			return Result{}, ErrInterrupted
		}
	}
	if ctx.Err() != nil {
		keepStaging = true
		return Result{}, ErrInterrupted
	}
	if err = os.Rename(staging, target); err != nil {
		if _, statErr := os.Stat(target); statErr == nil {
			return verify(target, environmentID, true)
		}
		return Result{}, ErrFenced
	}
	if err = syncDirectory(filepath.Dir(target)); err != nil {
		return Result{}, err
	}
	return Result{Generation: recoveryv1.Generation,
		TrustBundleSHA256: set.bundle.SHA256()}, nil
}

func writeSynced(path string, document []byte, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return ErrFenced
	}
	if _, err = file.Write(document); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		return ErrInterrupted
	}
	return nil
}
