package recoverysigning

import (
	"context"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"time"
)

type dependencies struct {
	random        io.Reader
	now           func() time.Time
	beforePublish func() error
}

func Reconcile(
	ctx context.Context,
	directory string,
	environmentID string,
) (Result, error) {
	return reconcile(ctx, directory, environmentID, dependencies{
		random: rand.Reader,
		now: func() time.Time {
			return time.Now().UTC().Truncate(time.Microsecond)
		},
	})
}

func reconcile(
	ctx context.Context,
	directory string,
	environmentID string,
	deps dependencies,
) (Result, error) {
	if ctx == nil || ctx.Err() != nil || !validTarget(directory) ||
		deps.random == nil || deps.now == nil {
		return Result{}, ErrInvalid
	}
	info, err := os.Lstat(directory)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return Result{}, ErrFenced
		}
		return verify(directory, environmentID, true)
	}
	if !os.IsNotExist(err) {
		return Result{}, ErrFenced
	}
	if err := prepareStaging(directory + ".staging"); err != nil {
		return Result{}, err
	}
	generatedAt := deps.now()
	if generatedAt.IsZero() || generatedAt.Location() != time.UTC ||
		generatedAt.Nanosecond()%1_000 != 0 {
		return Result{}, ErrInvalid
	}
	return generateAndPublish(ctx, directory, environmentID, generatedAt, deps)
}

func validTarget(directory string) bool {
	return filepath.IsAbs(directory) && filepath.Clean(directory) == directory &&
		filepath.Base(directory) != "." && filepath.Base(directory) != string(filepath.Separator)
}
