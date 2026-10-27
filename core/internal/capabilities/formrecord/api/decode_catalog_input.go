package recordkernel

type DecodeCatalogInput struct {
	CanonicalJSON  []byte
	ExpectedSHA256 [32]byte
	Binding        CatalogBinding
}
