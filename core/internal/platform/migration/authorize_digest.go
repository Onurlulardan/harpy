package migration

import "fmt"

func (registry *ReleaseRegistry) AuthorizeDigest(digest string) error {
	if registry == nil {
		return fmt.Errorf("migration release registry is not loaded")
	}
	canonical, exists := registry.manifests[digest]
	if !exists || len(canonical) == 0 {
		return fmt.Errorf("migration manifest digest is not embedded in this binary")
	}
	return nil
}
