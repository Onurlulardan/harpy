package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *graphRuntime) programs(
	ctx context.Context,
	scope transaction.TenantScope,
) (execution.ProgramStage, error) {
	if runtime == nil || runtime.barrier == nil {
		return execution.ProgramStage{}, ErrContract
	}
	rootCatalog, err := runtime.batch.record.Catalog()
	if err != nil {
		return execution.ProgramStage{}, ErrContract
	}
	root, err := mintProgram(
		ctx, scope, rootCatalog,
		func(programContext context.Context, driver execution.SnapshotDriver) error {
			return runtime.runRootProgram(programContext, driver)
		},
	)
	if err != nil {
		return execution.ProgramStage{}, ErrContract
	}
	programs := make([]execution.Program, 1, len(runtime.nodes)+1)
	programs[0] = root
	for index := range runtime.nodes {
		current := index
		catalog, err := runtime.nodes[current].compiled.Catalog()
		if err != nil {
			return execution.ProgramStage{}, ErrContract
		}
		program, err := mintProgram(
			ctx, scope, catalog,
			func(programContext context.Context, driver execution.SnapshotDriver) error {
				return runtime.runNodeProgram(programContext, driver, current)
			},
		)
		if err != nil {
			return execution.ProgramStage{}, ErrContract
		}
		programs = append(programs, program)
	}
	stage, err := execution.NewProgramStage(programs)
	if err != nil {
		return execution.ProgramStage{}, ErrContract
	}
	return stage, nil
}
