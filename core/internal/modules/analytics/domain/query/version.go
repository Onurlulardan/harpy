package query

import "bytes"

type Version struct {
	id       VersionID
	queryID  ID
	number   uint64
	revision uint64
	state    VersionState
	document [32]byte
	manifest [32]byte
}

func RehydrateVersion(
	id VersionID, queryID ID, number, revision uint64, state VersionState,
	document, manifest [32]byte,
) (Version, error) {
	if !id.Valid() || !queryID.Valid() || number == 0 || revision == 0 || !state.Valid() ||
		bytes.Equal(document[:], make([]byte, 32)) || bytes.Equal(manifest[:], make([]byte, 32)) {
		return Version{}, ErrInvalid
	}
	return Version{id: id, queryID: queryID, number: number, revision: revision,
		state: state, document: document, manifest: manifest}, nil
}

func (value Version) ID() VersionID            { return value.id }
func (value Version) QueryID() ID              { return value.queryID }
func (value Version) Number() uint64           { return value.number }
func (value Version) Revision() uint64         { return value.revision }
func (value Version) State() VersionState      { return value.state }
func (value Version) DocumentSHA256() [32]byte { return value.document }
func (value Version) ManifestSHA256() [32]byte { return value.manifest }

func (value Version) Publish(expected uint64) (Version, error) {
	if value.state != VersionDraft || expected == 0 || expected != value.revision {
		return Version{}, ErrConflict
	}
	value.state = VersionPublished
	value.revision++
	return value, nil
}
