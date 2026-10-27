package iam

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (application *authHTTPApplication) deviceEvidence(
	ctx context.Context,
) authentication.DeviceEvidence {
	ip, valid := authHTTPPeerIP(ctx)
	if !valid {
		return authentication.DeviceEvidence{}
	}
	keyID, key, err := application.network.Active()
	if err != nil {
		return authentication.DeviceEvidence{}
	}
	mac := hmac.New(sha256.New, key[:])
	for _, part := range []string{"harpy:iam:session-network:v1", ip} {
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(part)))
		_, _ = mac.Write(size[:])
		_, _ = mac.Write([]byte(part))
	}
	clear(key[:])
	var ipHMAC [32]byte
	copy(ipHMAC[:], mac.Sum(nil))
	return authentication.DeviceEvidence{
		Kind: "unknown", BrowserFamily: "unknown", OSFamily: "unknown",
		IPKeyID: keyID, IPHMAC: ipHMAC,
	}
}
