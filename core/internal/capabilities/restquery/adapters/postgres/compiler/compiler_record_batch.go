package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

// CompileRecordBatch atomically mints root and requested exact-count statements.
func (compiler Compiler) CompileRecordBatch(
	plan resolved.AuthorizedPlan,
) (RecordBatch, error) {
	if compiler.Validate() != nil || plan.Validate() != nil {
		return RecordBatch{}, ErrContract
	}
	binding, count, err := recordPlanBinding(plan)
	if err != nil {
		return RecordBatch{}, err
	}
	input, err := compiler.extractRecord(plan, count)
	if err != nil || input.binding != binding {
		return RecordBatch{}, ErrContract
	}
	record, err := compileRecord(input)
	if err != nil {
		return RecordBatch{}, err
	}
	var exact ExactCountStatement
	if count == api.CountModeExact {
		exact, err = compileExactCount(input)
		if err != nil {
			return RecordBatch{}, err
		}
	}
	return newRecordBatch(record, exact, count, binding)
}
