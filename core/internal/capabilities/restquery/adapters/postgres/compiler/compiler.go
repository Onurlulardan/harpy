package compiler

import (
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type Compiler struct {
	codec     cursor.Codec
	validated bool
}

func New(codec cursor.Codec) (Compiler, error) {
	compiler := Compiler{codec: codec, validated: true}
	if err := compiler.Validate(); err != nil {
		return Compiler{}, err
	}
	return compiler, nil
}

func (compiler Compiler) Validate() error {
	if !compiler.validated || compiler.codec.Validate() != nil {
		return ErrContract
	}
	return nil
}

// CompileRecord compiles only the root record-page statement.
func (compiler Compiler) CompileRecord(
	plan resolved.AuthorizedPlan,
) (RecordStatement, error) {
	if compiler.Validate() != nil || plan.Validate() != nil {
		return RecordStatement{}, ErrContract
	}
	input, err := compiler.extractRecord(plan, api.CountModeNone)
	if err != nil {
		return RecordStatement{}, err
	}
	return compileRecord(input)
}

func (Compiler) String() string                          { return compilerRedaction }
func (Compiler) GoString() string                        { return compilerRedaction }
func (Compiler) LogValue() slog.Value                    { return slog.StringValue(compilerRedaction) }
func (compiler Compiler) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Compiler) MarshalJSON() ([]byte, error)            { return rejectJSON() }
