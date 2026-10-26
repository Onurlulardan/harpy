package migratoroutput

import (
	"encoding/json"
	"fmt"
	"io"
)

func (result Result) Write(output io.Writer) error {
	if output == nil || result.value.Validate() != nil {
		return fmt.Errorf("migrator command result is invalid")
	}
	document, err := json.Marshal(result.value)
	if err != nil {
		return fmt.Errorf("migrator command result could not be encoded")
	}
	document = append(document, '\n')
	written, err := output.Write(document)
	if err != nil || written != len(document) {
		return fmt.Errorf("migrator command result could not be written")
	}
	return nil
}
