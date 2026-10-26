package effectexecution_test

import (
	"context"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

type providerStub struct {
	identity effectexecution.Identity
	handler  func(effectsv1.Request) effectexecution.Result
}

func (provider providerStub) Identity() effectexecution.Identity { return provider.identity }

func (provider providerStub) Execute(
	_ context.Context,
	request effectsv1.Request,
) effectexecution.Result {
	if provider.handler != nil {
		return provider.handler(request)
	}
	return effectexecution.Unknown()
}
