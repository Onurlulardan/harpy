package environmentoperator

import (
	"crypto/sha256"
	"strconv"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	"github.com/google/uuid"
)

const stableNamespace = "harpy-v2/environment-initialize/v1"

func stableID(manifest installationv1.Manifest, purpose string) uuid.UUID {
	hash := sha256.New()
	_, _ = hash.Write([]byte(stableNamespace))
	_, _ = hash.Write([]byte{0})
	environment := manifest.EnvironmentID()
	_, _ = hash.Write(environment[:])
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(manifest.RequestSHA256()))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(purpose))
	value := uuid.UUID(hash.Sum(nil)[:16])
	value[6] = value[6]&0x0f | 0x70
	value[8] = value[8]&0x3f | 0x80
	return value
}

func attemptID(manifest installationv1.Manifest, revision int64) uuid.UUID {
	return stableID(manifest, "attempt/"+strconv.FormatInt(revision, 10))
}
