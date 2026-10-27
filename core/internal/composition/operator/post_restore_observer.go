package operator

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

const (
	postRestoreObserverTimeout = 30 * time.Minute
	postRestoreObserverPoll    = 2 * time.Second
)

func runPostRestoreObserver(
	ctx context.Context,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	settings, err := config.LoadPostRestoreReactivationObserver()
	if err != nil {
		writeFailure(stderr, "configuration", processcode.InvalidStartup)
		return processcode.InvalidStartup
	}
	input, err := iam.ReadPostRestoreAwait(settings.Input().Path().Value())
	if err != nil {
		writeFailure(stderr, "input", processcode.InvalidStartup)
		return processcode.InvalidStartup
	}
	pool, err := openPostRestoreObserverPool(ctx, settings)
	if err != nil {
		writeFailure(stderr, "database", processcode.DependencyPreflight)
		return processcode.DependencyPreflight
	}
	defer pool.Close()
	observer, err := iam.NewPostRestoreObserver(pool)
	if err != nil {
		writeFailure(stderr, "database", processcode.DependencyPreflight)
		return processcode.DependencyPreflight
	}
	wait, cancel := context.WithTimeout(ctx, postRestoreObserverTimeout)
	defer cancel()
	for {
		observation, observeErr := observer.ObserveReactivation(
			wait,
			input.EnvironmentID, input.RestoreOperationID,
		)
		if observeErr != nil {
			code := processcode.OperatorResumable
			if postgres.ErrorHasSQLState(observeErr, "42501", "P0002") {
				code = processcode.OperatorTerminal
			}
			writeFailure(stderr, "observe", code)
			return code
		}
		if observation.State == "reactivated" {
			err = writePostRestoreObserved(
				stdout, input.EnvironmentID, input.RestoreOperationID,
				*observation.ReceiptSHA256, *observation.ReactivatedAt,
			)
			if err != nil {
				writeFailure(stderr, "result", processcode.OperatorTerminal)
				return processcode.OperatorTerminal
			}
			return processcode.Clean
		}
		timer := time.NewTimer(postRestoreObserverPoll)
		select {
		case <-wait.Done():
			timer.Stop()
			code := processcode.OperatorTerminal
			if ctx.Err() != nil {
				code = processcode.OperatorResumable
			}
			writeFailure(stderr, "wait", code)
			return code
		case <-timer.C:
		}
	}
}

func openPostRestoreObserverPool(
	ctx context.Context,
	settings config.PostRestoreReactivationObserver,
) (*postgres.Pool, error) {
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		return nil, err
	}
	embedded, err := schemacontract.LoadEmbedded(registry)
	if err != nil {
		return nil, err
	}
	runtime := settings.Runtime()
	connect, cancel := context.WithTimeout(
		ctx, runtime.Postgres().ConnectTimeout()+5*time.Second,
	)
	defer cancel()
	pool, err := postgres.Open(
		connect, runtime.Postgres(), postgres.ProfilePostRestoreObserver, embedded,
	)
	if err != nil {
		return nil, fmt.Errorf("open post-restore observer database: %w", err)
	}
	return pool, nil
}
