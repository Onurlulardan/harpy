package config

type CutoverAdoptionArtifacts struct {
	leaseManifestPath      AbsolutePath
	leaseSignaturePath     AbsolutePath
	trustRootPath          AbsolutePath
	sessionSealReceiptPath AbsolutePath
	batchPath              AbsolutePath
}

func loadCutoverAdoptionArtifacts(values environmentValues) (CutoverAdoptionArtifacts, error) {
	keys := []string{
		"HARPY_CUTOVER_LEASE_MANIFEST_PATH",
		"HARPY_CUTOVER_LEASE_SIGNATURE_PATH",
		"HARPY_CUTOVER_TRUST_ROOT_PATH",
		"HARPY_CUTOVER_SESSION_SEAL_RECEIPT_PATH",
		"HARPY_CUTOVER_BATCH_PATH",
	}
	paths, err := loadDistinctCutoverPaths(values, keys)
	if err != nil {
		return CutoverAdoptionArtifacts{}, err
	}
	return CutoverAdoptionArtifacts{
		leaseManifestPath: paths[0], leaseSignaturePath: paths[1], trustRootPath: paths[2],
		sessionSealReceiptPath: paths[3], batchPath: paths[4],
	}, nil
}

func (a CutoverAdoptionArtifacts) LeaseManifestPath() AbsolutePath  { return a.leaseManifestPath }
func (a CutoverAdoptionArtifacts) LeaseSignaturePath() AbsolutePath { return a.leaseSignaturePath }
func (a CutoverAdoptionArtifacts) TrustRootPath() AbsolutePath      { return a.trustRootPath }
func (a CutoverAdoptionArtifacts) SessionSealReceiptPath() AbsolutePath {
	return a.sessionSealReceiptPath
}
func (a CutoverAdoptionArtifacts) BatchPath() AbsolutePath { return a.batchPath }

func loadDistinctCutoverPaths(
	values environmentValues,
	keys []string,
) ([]AbsolutePath, error) {
	paths := make([]AbsolutePath, 0, len(keys))
	for _, key := range keys {
		path, err := loadAbsolutePath(values, key)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	if duplicateReleasePath(paths...) {
		return nil, configError("HARPY_CUTOVER_*_PATH", "cutover artifact paths must be distinct")
	}
	return paths, nil
}
