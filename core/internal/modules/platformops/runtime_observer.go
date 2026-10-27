package platformops

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

type runtimeTraceObserver struct {
	environment HTTPEnvironmentReader
	traces      runtimeTraceRecorder
	entropy     io.Reader
}

type runtimeTraceRecorder interface {
	RecordTrace(context.Context, platformopsapi.TraceObservation) error
}

func newRuntimeTraceObserver(
	environment HTTPEnvironmentReader,
	traces runtimeTraceRecorder,
) (*runtimeTraceObserver, error) {
	if environment == nil || traces == nil {
		return nil, fmt.Errorf("PlatformOps trace observer dependencies are required")
	}
	return &runtimeTraceObserver{environment: environment, traces: traces,
		entropy: rand.Reader}, nil
}

func (observer *runtimeTraceObserver) Observe(
	ctx context.Context,
	request httpedge.RequestObservation,
) error {
	if observer == nil || observer.environment == nil || observer.traces == nil ||
		observer.entropy == nil || ctx == nil || ctx.Err() != nil {
		return fmt.Errorf("PlatformOps trace observer is unavailable")
	}
	if strings.HasPrefix(request.RoutePattern, "/api/platform/operations") {
		return nil
	}
	environment, err := observer.environment.Current(ctx)
	if err != nil {
		return err
	}
	spanID, err := observer.nextSpanID()
	if err != nil {
		return err
	}
	level := "info"
	if request.StatusCode >= 500 {
		level = "error"
	} else if request.StatusCode >= 400 {
		level = "warn"
	}
	return observer.traces.RecordTrace(ctx, platformopsapi.TraceObservation{
		EnvironmentID: environment.EnvironmentID,
		CorrelationID: request.CorrelationID.String(), SpanID: spanID,
		ServiceName: "core-api", Operation: request.Method + " " + request.RoutePattern,
		SpanKind: "server", Method: request.Method, Path: request.RoutePattern,
		StatusCode: request.StatusCode, StartedAt: request.StartedAt, EndedAt: request.EndedAt,
		DurationMillis: float64(request.EndedAt.Sub(request.StartedAt).Microseconds()) / 1000,
		Level:          level,
	})
}

func (observer *runtimeTraceObserver) nextSpanID() (string, error) {
	var value [8]byte
	if _, err := io.ReadFull(observer.entropy, value[:]); err != nil {
		return "", fmt.Errorf("generate PlatformOps span ID: %w", err)
	}
	return hex.EncodeToString(value[:]), nil
}
