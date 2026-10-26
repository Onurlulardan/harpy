package executor

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

type Executor struct {
	registry         execution.Registry
	postgresCompiler compiler.Compiler
	codec            cursor.Codec
	policy           planner.Policy
	valid            bool
}

func New(
	registry execution.Registry,
	codec cursor.Codec,
	policy planner.Policy,
) (Executor, error) {
	postgresCompiler, err := compiler.New(codec)
	if err != nil {
		return Executor{}, ErrContract
	}
	executor := Executor{
		registry: registry, postgresCompiler: postgresCompiler,
		codec: codec, policy: policy, valid: true,
	}
	if executor.Validate() != nil {
		return Executor{}, ErrContract
	}
	return executor, nil
}

func (executor Executor) Validate() error {
	if !executor.valid || executor.registry.Validate() != nil ||
		executor.postgresCompiler.Validate() != nil || executor.codec.Validate() != nil ||
		validatePolicy(executor.policy) != nil {
		return ErrContract
	}
	return nil
}

func (Executor) String() string   { return redacted }
func (Executor) GoString() string { return redacted }
func (Executor) LogValue() slog.Value {
	return slog.StringValue(redacted)
}
func (executor Executor) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Executor) MarshalJSON() ([]byte, error)            { return nil, ErrDirectMarshal }

func validatePolicy(policy planner.Policy) error {
	if _, err := planner.NewTracker(policy); err != nil {
		return err
	}
	limits := policy.Limits()
	if limits.InteractiveTimeout%time.Millisecond != 0 ||
		limits.ExactCountTimeout%time.Millisecond != 0 {
		return ErrContract
	}
	return nil
}
