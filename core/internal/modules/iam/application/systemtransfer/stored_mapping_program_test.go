package systemtransfer

import (
	"errors"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestBuildStoredMappingProgramRejectsUnknownResource(t *testing.T) {
	_, err := BuildStoredMappingProgram(publicresource.Resource("unknown"), []byte(`{}`))
	if !errors.Is(err, ErrInvalidMapping) {
		t.Fatalf("BuildStoredMappingProgram() error = %v", err)
	}
}
