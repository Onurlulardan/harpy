package schemacontract

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
)

func LoadStatic() (EmbeddedContract, error) {
	artifact, err := binarycontract.LoadEmbedded()
	if err != nil {
		return EmbeddedContract{}, err
	}
	if !artifact.IsVerified() {
		return EmbeddedContract{}, fmt.Errorf("embedded binary schema contract is not verified")
	}
	return EmbeddedContract{artifact: artifact, authorized: true}, nil
}
