package executor

import (
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

// GlobalExecutor is the one-owner global snapshot execution surface.
type GlobalExecutor struct {
	registry         execution.GlobalRegistry
	postgresCompiler compiler.Compiler
	codec            cursor.Codec
	policy           planner.Policy
	valid            bool
}

func NewGlobal(
	registry execution.GlobalRegistry,
	codec cursor.Codec,
	policy planner.Policy,
) (GlobalExecutor, error) {
	postgresCompiler, err := compiler.New(codec)
	if err != nil {
		return GlobalExecutor{}, ErrContract
	}
	executor := GlobalExecutor{
		registry: registry, postgresCompiler: postgresCompiler,
		codec: codec, policy: policy, valid: true,
	}
	if executor.Validate() != nil {
		return GlobalExecutor{}, ErrContract
	}
	return executor, nil
}

func (executor GlobalExecutor) Validate() error {
	if !executor.valid || executor.registry.Validate() != nil ||
		executor.postgresCompiler.Validate() != nil || executor.codec.Validate() != nil ||
		validatePolicy(executor.policy) != nil {
		return ErrContract
	}
	return nil
}

func (GlobalExecutor) String() string   { return redacted }
func (GlobalExecutor) GoString() string { return redacted }
func (GlobalExecutor) LogValue() slog.Value {
	return slog.StringValue(redacted)
}
func (executor GlobalExecutor) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (GlobalExecutor) MarshalJSON() ([]byte, error)            { return nil, ErrDirectMarshal }
