package runtime

// NewExact builds the canonical engine for an owner whose application resolves
// an exact immutable CatalogSet before invoking ExecuteCatalog.
func NewExact(
	owner string,
	policy PolicyPort,
	ownerPort OwnerPort,
	keys CursorKeys,
) (*Engine, error) {
	return newEngine(owner, nil, policy, ownerPort, keys)
}
