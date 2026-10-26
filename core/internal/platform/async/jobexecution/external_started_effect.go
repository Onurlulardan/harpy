package jobexecution

import (
	"fmt"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

// StartedEffect is an opaque bridge from the tenant-bound pre-effect
// transaction to the provider invocation and worker-owned finalization.
type StartedEffect struct {
	permit  asyncpostgres.ExternalEffectPermit
	request effectsv1.Request
}

func NewStartedEffect(
	permit asyncpostgres.ExternalEffectPermit,
) (StartedEffect, error) {
	request := permit.Request()
	if request.Validate() != nil || permit.StartedAt().IsZero() {
		return StartedEffect{}, fmt.Errorf("%w: permit evidence", ErrExternalEffectStart)
	}
	return StartedEffect{permit: permit, request: request}, nil
}

func (started StartedEffect) Request() effectsv1.Request {
	document, err := started.request.MarshalJSON()
	if err != nil {
		return effectsv1.Request{}
	}
	request, err := effectsv1.DecodeRequest(document)
	if err != nil {
		return effectsv1.Request{}
	}
	return request
}

func (started StartedEffect) validFor(
	spec asyncpostgres.ExternalEffectRequestSpec,
) bool {
	return started.request.Validate() == nil &&
		started.request.Provider == spec.Provider() &&
		started.request.Operation == spec.Operation()
}
