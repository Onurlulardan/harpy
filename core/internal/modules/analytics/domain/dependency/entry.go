package dependency

import "bytes"

type Entry struct {
	ownerDomain       string
	resourceKind      string
	resourceID        [16]byte
	resourceVersionID [16]byte
	catalogVersion    uint64
	queryContract     uint64
	policyVersion     uint64
	dataRevision      string
	digest            [32]byte
}

func NewEntry(
	ownerDomain, resourceKind string, resourceID, resourceVersionID [16]byte,
	catalogVersion, queryContract, policyVersion uint64,
	dataRevision string, digest [32]byte,
) (Entry, error) {
	if !allowedPhase8Owner(ownerDomain) || resourceKind == "" ||
		!validID(resourceID) || !validID(resourceVersionID) || catalogVersion == 0 ||
		queryContract == 0 || policyVersion == 0 || zeroDigest(digest) ||
		len(resourceKind) > 64 || len(dataRevision) > 128 {
		return Entry{}, ErrInvalid
	}
	return Entry{ownerDomain: ownerDomain, resourceKind: resourceKind,
		resourceID: resourceID, resourceVersionID: resourceVersionID,
		catalogVersion: catalogVersion, queryContract: queryContract,
		policyVersion: policyVersion, dataRevision: dataRevision, digest: digest}, nil
}

func (entry Entry) OwnerDomain() string            { return entry.ownerDomain }
func (entry Entry) ResourceKind() string           { return entry.resourceKind }
func (entry Entry) ResourceID() [16]byte           { return entry.resourceID }
func (entry Entry) ResourceVersionID() [16]byte    { return entry.resourceVersionID }
func (entry Entry) CatalogVersion() uint64         { return entry.catalogVersion }
func (entry Entry) QueryContractVersion() uint64   { return entry.queryContract }
func (entry Entry) PolicyVersion() uint64          { return entry.policyVersion }
func (entry Entry) DataRevisionCapability() string { return entry.dataRevision }
func (entry Entry) SHA256() [32]byte               { return entry.digest }

func allowedPhase8Owner(value string) bool { return value == "forms" || value == "iam" }
func validID(value [16]byte) bool          { var zero [16]byte; return value != zero }
func zeroDigest(value [32]byte) bool       { return bytes.Equal(value[:], make([]byte, 32)) }
