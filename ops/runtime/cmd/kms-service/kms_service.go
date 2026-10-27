package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"regexp"
	"strings"

	"cloud.google.com/go/kms/apiv1/kmspb"
)

const (
	aesKeyBytes    = 32
	gcmNonceBytes  = 12
	gcmTagBytes    = 16
	maxPayloadSize = 64 * 1024
)

var keyVersionPattern = regexp.MustCompile(
	`^projects/([a-z][a-z0-9-]{4,28}[a-z0-9]|[0-9]{6,30})/` +
		`locations/[a-z0-9][a-z0-9-]{0,62}/keyRings/[A-Za-z0-9_-]{1,63}/` +
		`cryptoKeys/[A-Za-z0-9_-]{1,63}/cryptoKeyVersions/[1-9][0-9]{0,18}$`,
)

type kmsService struct {
	kmspb.UnimplementedKeyManagementServiceServer
	master        [aesKeyBytes]byte
	projectPrefix string
}

func newKMSService(key []byte, profiles ...string) *kmsService {
	service := &kmsService{}
	copy(service.master[:], key)
	profile := "local"
	if len(profiles) == 1 {
		profile = profiles[0]
	}
	service.projectPrefix = "projects/harpy-" + profile + "/"
	return service
}

func (service *kmsService) validResource(name string) bool {
	return strings.HasPrefix(name, service.projectPrefix) &&
		keyVersionPattern.MatchString(name)
}

func (service *kmsService) keyFor(name string) [aesKeyBytes]byte {
	deriver := hmac.New(sha256.New, service.master[:])
	_, _ = deriver.Write([]byte("harpy-kms-resource-v1\x00"))
	_, _ = deriver.Write([]byte(name))
	var key [aesKeyBytes]byte
	copy(key[:], deriver.Sum(nil))
	return key
}
