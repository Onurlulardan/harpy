package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	usagemodule "github.com/Onurlulardan/harpy/core/internal/modules/usage"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestattribution"
)

var (
	errUsageObservationConflict = errors.New("Usage request attribution conflicts")
	errUsageObservationInvalid  = errors.New("Usage request observation is invalid")
)

type usageRequestRecorder interface {
	Record(context.Context, usagemodule.RecordRequest) (usagemodule.RecordResult, error)
}

func newUsageRequestObserver(
	environmentID [16]byte, recorder usageRequestRecorder,
) (httpedge.RequestObserver, error) {
	if environmentID == [16]byte{} || recorder == nil {
		return nil, fmt.Errorf("Usage request observer dependencies are required")
	}
	return func(ctx context.Context, observation httpedge.RequestObservation) error {
		if ctx == nil || ctx.Err() != nil {
			return errUsageObservationInvalid
		}
		if usageObservationRoute(observation.RoutePattern) {
			return nil
		}
		switch observation.AttributionState {
		case requestattribution.StatusUnattributed:
			return nil
		case requestattribution.StatusConflict:
			return errUsageObservationConflict
		case requestattribution.StatusVerified:
		default:
			return errUsageObservationInvalid
		}
		duration := observation.EndedAt.Sub(observation.StartedAt).Microseconds()
		if !observation.Attribution.Valid() || observation.SourceIdentity == [32]byte{} ||
			observation.Method == "" || observation.RoutePattern == "" ||
			observation.StatusCode < 100 || observation.StatusCode > 599 || duration < 0 {
			return errUsageObservationInvalid
		}
		_, err := recorder.Record(ctx, usagemodule.RecordRequest{
			EnvironmentID:  environmentID,
			TenantID:       observation.Attribution.TenantID,
			ActorID:        observation.Attribution.ActorID,
			SourceIdentity: observation.SourceIdentity,
			Method:         observation.Method, RoutePattern: observation.RoutePattern,
			StatusCode: observation.StatusCode, DurationMicros: duration,
			OccurredAt: observation.StartedAt.UTC().Truncate(time.Microsecond),
		})
		if err != nil {
			return fmt.Errorf("record Usage request observation: %w", err)
		}
		return nil
	}, nil
}

func usageObservationRoute(pattern string) bool {
	return pattern == "/api/usage-analytics" ||
		strings.HasPrefix(pattern, "/api/usage-analytics/")
}
