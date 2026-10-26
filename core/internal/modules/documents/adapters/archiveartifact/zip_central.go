package archiveartifact

import (
	"io"
	"math"
)

func writeCentral(target io.Writer, entry centralEntry) error {
	name := []byte(entry.name)
	version, offset, extraBytes := uint16(20), uint32(entry.localOffset), uint16(0)
	if entry.zip64Offset {
		version, offset, extraBytes = 45, math.MaxUint32, uint16(zipCentralZip64)
	}
	if err := write32(target, 0x02014b50); err != nil {
		return err
	}
	if err := write16(target, 20, version, zipFlags, 0, 0, 0); err != nil {
		return err
	}
	if err := write32(target, entry.crc, entry.bytes, entry.bytes); err != nil {
		return err
	}
	if err := write16(target, uint16(len(name)), extraBytes, 0, 0, 0); err != nil {
		return err
	}
	if err := write32(target, 0, offset); err != nil {
		return err
	}
	if _, err := target.Write(name); err != nil {
		return err
	}
	if entry.zip64Offset {
		if err := write16(target, 0x0001, 8); err != nil {
			return err
		}
		return write64(target, entry.localOffset)
	}
	return nil
}

func writeZIP64End(target *countingWriter, count int, centralBytes,
	centralOffset uint64,
) error {
	recordOffset := target.count
	if err := write32(target, 0x06064b50); err != nil {
		return err
	}
	if err := write64(target, 44); err != nil {
		return err
	}
	if err := write16(target, 45, 45); err != nil {
		return err
	}
	if err := write32(target, 0, 0); err != nil {
		return err
	}
	if err := write64(target, uint64(count), uint64(count),
		centralBytes, centralOffset); err != nil {
		return err
	}
	if err := write32(target, 0x07064b50, 0); err != nil {
		return err
	}
	if err := write64(target, recordOffset); err != nil {
		return err
	}
	return write32(target, 1)
}

func writeEnd(target io.Writer, count int, centralBytes, centralOffset uint64) error {
	if err := write32(target, 0x06054b50); err != nil {
		return err
	}
	if err := write16(target, 0, 0, uint16(count), uint16(count)); err != nil {
		return err
	}
	size, offset := uint32(centralBytes), uint32(centralOffset)
	if centralBytes > math.MaxUint32 {
		size = math.MaxUint32
	}
	if centralOffset > math.MaxUint32 {
		offset = math.MaxUint32
	}
	if err := write32(target, size, offset); err != nil {
		return err
	}
	return write16(target, 0)
}
