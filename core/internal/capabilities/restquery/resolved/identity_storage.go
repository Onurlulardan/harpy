package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func writeResourceStorage(writer *identityWriter, storage api.ResourceStorage) {
	fingerprint, err := storage.FingerprintSHA256()
	if err != nil {
		writer.err = fmt.Errorf("resolved REST Query resource storage is invalid")
		return
	}
	writer.writeBytes(fingerprint[:])
}

func writeRelationStorage(writer *identityWriter, storage api.RelationStorage) {
	fingerprint, err := storage.FingerprintSHA256()
	if err != nil {
		writer.err = fmt.Errorf("resolved REST Query relation storage is invalid")
		return
	}
	writer.writeBytes(fingerprint[:])
}
