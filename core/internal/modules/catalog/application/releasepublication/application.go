package releasepublication

import (
	"crypto/ed25519"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

type KeyResolver interface {
	Resolve(keyID string) (ed25519.PublicKey, error)
}

type Application struct {
	registrar packageregistry.Registrar
	keys      KeyResolver
}

func NewApplication(
	registrar packageregistry.Registrar,
	keys KeyResolver,
) (*Application, error) {
	if registrar == nil || keys == nil {
		return nil, fmt.Errorf("catalog release publication dependencies are required")
	}
	return &Application{registrar: registrar, keys: keys}, nil
}
