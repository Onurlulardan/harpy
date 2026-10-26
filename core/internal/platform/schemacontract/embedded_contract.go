package schemacontract

import (
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
)

type EmbeddedContract struct {
	artifact   binarycontract.Artifact
	authorized bool
}

func (contract EmbeddedContract) Value() runtimev1.BinarySchemaContract {
	if !contract.authorized {
		return runtimev1.BinarySchemaContract{}
	}
	return contract.artifact.Value()
}

func (contract EmbeddedContract) DigestSHA256() string {
	if !contract.authorized {
		return ""
	}
	return contract.artifact.DigestSHA256()
}

func (contract EmbeddedContract) IsVerified() bool {
	return contract.authorized && contract.artifact.IsVerified()
}
