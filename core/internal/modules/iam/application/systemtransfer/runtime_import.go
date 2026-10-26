package systemtransfer

import tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"

type ImportRuntime struct {
	loader  AttemptLoader
	sources ExactImportSourceOpener
	batches *Worker
	limits  tabulartransfer.Limits
}

func NewSystemImportRuntime(
	loader AttemptLoader,
	sources ExactImportSourceOpener,
	batches *Worker,
) (*ImportRuntime, error) {
	if nilDependency(loader) || nilDependency(sources) || batches == nil {
		return nil, ErrDependencies
	}
	limits := tabulartransfer.DefaultLimits()
	if limits.Validate() != nil {
		return nil, ErrDependencies
	}
	return &ImportRuntime{
		loader: loader, sources: sources, batches: batches, limits: limits,
	}, nil
}
