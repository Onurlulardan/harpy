package dependency

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"sort"
)

type Manifest struct {
	entries []Entry
	digest  [32]byte
}

func NewManifest(entries []Entry, revisionBound bool) (Manifest, error) {
	if len(entries) == 0 || len(entries) > 32 {
		return Manifest{}, ErrInvalid
	}
	ordered := append([]Entry(nil), entries...)
	sort.Slice(ordered, func(i, j int) bool { return entryLess(ordered[i], ordered[j]) })
	hash := sha256.New()
	var previous string
	for _, entry := range ordered {
		identity := entryIdentity(entry)
		if identity == previous || (revisionBound && entry.dataRevision == "") {
			return Manifest{}, ErrInvalid
		}
		previous = identity
		writeEntry(hash, entry)
	}
	var digest [32]byte
	copy(digest[:], hash.Sum(nil))
	return Manifest{entries: ordered, digest: digest}, nil
}

func (manifest Manifest) Entries() []Entry { return append([]Entry(nil), manifest.entries...) }
func (manifest Manifest) SHA256() [32]byte { return manifest.digest }

func entryLess(left, right Entry) bool { return entryIdentity(left) < entryIdentity(right) }

func entryIdentity(entry Entry) string {
	return entry.ownerDomain + "\x00" + entry.resourceKind + "\x00" +
		string(entry.resourceID[:]) + string(entry.resourceVersionID[:])
}

func writeEntry(hash interface{ Write([]byte) (int, error) }, entry Entry) {
	for _, value := range [][]byte{[]byte(entry.ownerDomain), []byte(entry.resourceKind),
		entry.resourceID[:], entry.resourceVersionID[:], []byte(entry.dataRevision), entry.digest[:]} {
		length := make([]byte, 4)
		binary.BigEndian.PutUint32(length, uint32(len(value)))
		_, _ = hash.Write(length)
		_, _ = hash.Write(value)
	}
	versions := make([]byte, 24)
	binary.BigEndian.PutUint64(versions[0:8], entry.catalogVersion)
	binary.BigEndian.PutUint64(versions[8:16], entry.queryContract)
	binary.BigEndian.PutUint64(versions[16:24], entry.policyVersion)
	_, _ = hash.Write(versions)
}

func (manifest Manifest) Equal(other Manifest) bool {
	return bytes.Equal(manifest.digest[:], other.digest[:])
}
