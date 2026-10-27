package v1query

func ValidateManifestSnapshot(snapshot ManifestSnapshot) error {
	_, err := newManifest(snapshot)
	return err
}
