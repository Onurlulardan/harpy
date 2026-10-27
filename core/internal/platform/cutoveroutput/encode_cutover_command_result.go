package cutoveroutput

import (
	"encoding/json"
	"fmt"
	"io"
)

func (result Result) Encode(output io.Writer) error {
	if err := result.value.Validate(); err != nil {
		return fmt.Errorf("validate cutover command result: %w", err)
	}
	if err := json.NewEncoder(output).Encode(result.value); err != nil {
		return fmt.Errorf("encode cutover command result: %w", err)
	}
	return nil
}
