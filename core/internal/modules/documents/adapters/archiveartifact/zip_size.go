package archiveartifact

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

const (
	zipLocalHeader     = int64(30)
	zipDescriptor      = int64(16)
	zipCentralHeader   = int64(46)
	zipCentralZip64    = int64(12)
	zipEnd             = int64(22)
	zip64EndAndLocator = int64(76)
)

type zipLayoutEntry struct {
	name        string
	bytes       uint32
	localOffset uint64
	zip64Offset bool
}

func zipLayout(entries []archiveapp.Entry) ([]zipLayoutEntry, int64, error) {
	if len(entries) == 0 || len(entries) > domain.MaximumArchiveFiles {
		return nil, 0, archiveapp.ErrLimit
	}
	layout := make([]zipLayoutEntry, len(entries))
	var offset int64
	for index, entry := range entries {
		if entry.Order != uint32(index) || !entry.DocumentID.UUIDv7() ||
			!entry.VersionID.UUIDv7() || entry.PlaintextBytes < 1 ||
			entry.PlaintextBytes > domain.MaxUploadBytes ||
			entry.PlaintextSHA256 == [32]byte{} {
			return nil, 0, archiveapp.ErrInvalid
		}
		name := archiveFilename(index, entry.OriginalName)
		layout[index] = zipLayoutEntry{name: name, bytes: uint32(entry.PlaintextBytes),
			localOffset: uint64(offset), zip64Offset: offset > math.MaxUint32}
		offset += zipLocalHeader + int64(len([]byte(name))) +
			entry.PlaintextBytes + zipDescriptor
		if offset > domain.MaximumArchiveArtifactPlaintext {
			return nil, 0, archiveapp.ErrLimit
		}
	}
	centralOffset := offset
	for _, entry := range layout {
		extra := int64(0)
		if entry.zip64Offset {
			extra = zipCentralZip64
		}
		offset += zipCentralHeader + int64(len([]byte(entry.name))) + extra
	}
	if centralOffset > math.MaxUint32 || offset-centralOffset > math.MaxUint32 {
		offset += zip64EndAndLocator
	}
	offset += zipEnd
	if offset > domain.MaximumArchiveArtifactPlaintext {
		return nil, 0, archiveapp.ErrLimit
	}
	return layout, offset, nil
}

func zipSize(entries []archiveapp.Entry) (int64, error) {
	_, size, err := zipLayout(entries)
	return size, err
}

func archiveFilename(index int, original string) string {
	cleaned := strings.Map(func(value rune) rune {
		if value == '/' || value == '\\' || unicode.IsControl(value) {
			return '_'
		}
		return value
	}, strings.TrimSpace(original))
	cleaned = strings.Trim(cleaned, ". ")
	if cleaned == "" {
		cleaned = "document"
	}
	return fmt.Sprintf("%05d-%s", index+1, cleaned)
}
