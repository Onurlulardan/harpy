package testadapter

import (
	"context"
	"sync"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

type Handler func(effectsv1.Request) effectexecution.Result

type Adapter struct {
	identity effectexecution.Identity
	handler  Handler
	mutex    sync.Mutex
	requests []effectsv1.Request
}

func New(
	identity effectexecution.Identity,
	handler Handler,
) (*Adapter, error) {
	if handler == nil {
		return nil, effectexecution.ErrProviderResult
	}
	return &Adapter{identity: identity, handler: handler}, nil
}

func (adapter *Adapter) Identity() effectexecution.Identity { return adapter.identity }

func (adapter *Adapter) Execute(
	_ context.Context,
	request effectsv1.Request,
) effectexecution.Result {
	copy := cloneRequest(request)
	adapter.mutex.Lock()
	adapter.requests = append(adapter.requests, copy)
	adapter.mutex.Unlock()
	return adapter.handler(copy)
}

func (adapter *Adapter) Requests() []effectsv1.Request {
	adapter.mutex.Lock()
	defer adapter.mutex.Unlock()
	result := make([]effectsv1.Request, len(adapter.requests))
	for index, request := range adapter.requests {
		result[index] = cloneRequest(request)
	}
	return result
}

func cloneRequest(request effectsv1.Request) effectsv1.Request {
	document, err := request.MarshalJSON()
	if err != nil {
		return effectsv1.Request{}
	}
	copy, err := effectsv1.DecodeRequest(document)
	if err != nil {
		return effectsv1.Request{}
	}
	return copy
}
