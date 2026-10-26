package archiveartifact

import (
	"context"
	"crypto/sha256"
	"hash/crc32"
	"io"
	"math"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
)

const zipFlags = uint16(0x0808)

type centralEntry struct {
	zipLayoutEntry
	crc uint32
}

func writeZIP(ctx context.Context, target io.Writer, entries []archiveapp.Entry,
	contents ContentSource, refresh func(int) error,
) error {
	layout, expected, err := zipLayout(entries)
	if err != nil {
		return err
	}
	counter := &countingWriter{target: target}
	central := make([]centralEntry, len(entries))
	for index, entry := range entries {
		if err = refresh(index); err != nil {
			return err
		}
		central[index], err = writeZIPEntry(ctx, counter, layout[index], entry, contents)
		if err != nil {
			return err
		}
	}
	centralOffset := counter.count
	for _, entry := range central {
		if err = writeCentral(counter, entry); err != nil {
			return err
		}
	}
	centralBytes := counter.count - centralOffset
	if centralOffset > math.MaxUint32 || centralBytes > math.MaxUint32 {
		if err = writeZIP64End(counter, len(entries), centralBytes, centralOffset); err != nil {
			return err
		}
	}
	if err = writeEnd(counter, len(entries), centralBytes, centralOffset); err != nil {
		return err
	}
	if counter.count != uint64(expected) {
		return archiveapp.ErrUnavailable
	}
	return nil
}

func writeZIPEntry(ctx context.Context, target io.Writer, layout zipLayoutEntry,
	entry archiveapp.Entry, contents ContentSource,
) (centralEntry, error) {
	name := []byte(layout.name)
	if err := write32(target, 0x04034b50); err != nil {
		return centralEntry{}, err
	}
	if err := write16(target, 20, zipFlags, 0, 0, 0); err != nil {
		return centralEntry{}, err
	}
	if err := write32(target, 0, 0, 0); err != nil {
		return centralEntry{}, err
	}
	if err := write16(target, uint16(len(name)), 0); err != nil {
		return centralEntry{}, err
	}
	if _, err := target.Write(name); err != nil {
		return centralEntry{}, err
	}
	source, err := contents.Open(ctx, entry)
	if err != nil {
		return centralEntry{}, err
	}
	defer source.Close()
	hash := crc32.NewIEEE()
	contentHash := sha256.New()
	if _, err = io.CopyN(io.MultiWriter(target, hash, contentHash), source,
		entry.PlaintextBytes); err != nil {
		return centralEntry{}, err
	}
	var extra [1]byte
	if count, readErr := source.Read(extra[:]); count != 0 || readErr != io.EOF {
		return centralEntry{}, archiveapp.ErrUnavailable
	}
	var digest [32]byte
	copy(digest[:], contentHash.Sum(nil))
	if digest != entry.PlaintextSHA256 {
		return centralEntry{}, archiveapp.ErrUnavailable
	}
	if err = write32(target, 0x08074b50, hash.Sum32(), layout.bytes, layout.bytes); err != nil {
		return centralEntry{}, err
	}
	return centralEntry{zipLayoutEntry: layout, crc: hash.Sum32()}, nil
}

type countingWriter struct {
	target io.Writer
	count  uint64
}

func (writer *countingWriter) Write(value []byte) (int, error) {
	count, err := writer.target.Write(value)
	writer.count += uint64(count)
	return count, err
}
