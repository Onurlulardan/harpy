package httpedge

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/requestattribution"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type RequestObservation struct {
	CorrelationID    requestmeta.CorrelationID
	SourceIdentity   [32]byte
	Attribution      requestattribution.Identity
	AttributionState requestattribution.Status
	Method           string
	RoutePattern     string
	StatusCode       int
	StartedAt        time.Time
	EndedAt          time.Time
}

type RequestObserver func(context.Context, RequestObservation) error

func (observation RequestObservation) valid() bool {
	attributionValid := observation.AttributionState == requestattribution.StatusUnattributed ||
		observation.AttributionState == requestattribution.StatusConflict ||
		observation.AttributionState == requestattribution.StatusVerified &&
			observation.Attribution.Valid()
	return observation.CorrelationID != "" && observation.SourceIdentity != [32]byte{} &&
		attributionValid && observation.Method != "" &&
		observation.RoutePattern != "" && observation.StatusCode >= 100 &&
		observation.StatusCode <= 599 && !observation.StartedAt.IsZero() &&
		!observation.EndedAt.Before(observation.StartedAt)
}
