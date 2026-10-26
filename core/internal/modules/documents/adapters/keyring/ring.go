package keyring

import (
	"io"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/keywrap"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Release struct {
	ID      string
	Current bool
	Key     [32]byte
}

type Ring struct {
	current    string
	keys       map[string][32]byte
	randomness io.Reader
}

func New(releases []Release, randomness io.Reader) (*Ring, error) {
	if randomness == nil || len(releases) == 0 || len(releases) > 8 {
		return nil, uploadapp.ErrUnavailable
	}
	ring := &Ring{keys: make(map[string][32]byte, len(releases)), randomness: randomness}
	for _, release := range releases {
		if !validID(release.ID) || release.Key == [32]byte{} ||
			ring.keys[release.ID] != [32]byte{} || release.Current && ring.current != "" {
			return nil, uploadapp.ErrUnavailable
		}
		ring.keys[release.ID] = release.Key
		if release.Current {
			ring.current = release.ID
		}
	}
	if ring.current == "" {
		return nil, uploadapp.ErrUnavailable
	}
	return ring, nil
}

func validID(value string) bool {
	if value == "" || len(value) > 64 || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if character < 'a' || character > 'z' {
			if character < '0' || character > '9' {
				if character != '-' && character != '_' && character != '.' {
					return false
				}
			}
		}
	}
	return true
}

func wrapContext(tenantID, documentID, versionID domain.ID,
	keyID string,
) keywrap.Context {
	return keywrap.Context{TenantID: tenantID, DocumentID: documentID,
		VersionID: versionID, KeyID: keyID}
}

var _ uploadapp.Keyring = (*Ring)(nil)
