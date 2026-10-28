package main

import "github.com/Onurlulardan/harpy/internal/migrationbundle"

// schemaMigrationManifest is the setup filesystem producer. It uses the same
// select, manifest, and canonical bytes as the runtime fresh registry.
// History-supersession chains stay on runtime migration documents; they are
// not fields of the fresh manifest.
func schemaMigrationManifest(root string) ([]byte, error) {
	return migrationbundle.CanonicalFresh(root)
}
