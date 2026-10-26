package adoptionartifact

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type Copier struct {
	source   SourceStore
	metadata MetadataVerifier
	keyring  Keyring
	cipher   Cipher
	objects  ObjectStore
}

func New(source SourceStore, metadata MetadataVerifier, keyring Keyring,
	cipher Cipher, objects ObjectStore,
) (*Copier, error) {
	for _, dependency := range []any{source, metadata, keyring, cipher, objects} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents adoption artifact dependencies are invalid")
		}
	}
	return &Copier{source: source, metadata: metadata, keyring: keyring,
		cipher: cipher, objects: objects}, nil
}
