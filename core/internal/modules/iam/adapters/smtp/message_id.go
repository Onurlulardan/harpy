package smtp

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func (client *Client) StableMessageID(effectKey string) (string, error) {
	if client == nil || len(effectKey) < 1 || len(effectKey) > 512 {
		return "", ErrContract
	}
	separator := strings.LastIndexByte(client.settings.from, '@')
	if separator < 1 || separator == len(client.settings.from)-1 {
		return "", ErrContract
	}
	digest := sha256.Sum256([]byte("iam.invitation.smtp-message.v1\x00" + effectKey))
	return "invite." + hex.EncodeToString(digest[:]) + client.settings.from[separator:], nil
}
