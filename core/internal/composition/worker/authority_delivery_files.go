package worker

import (
	"crypto/rand"
	"fmt"
	"time"
)

type authorityWorkerClock struct{}

func (authorityWorkerClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

func newAuthorityDeliveryOwnerID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("create platform authority delivery owner: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		value[0:4],
		value[4:6],
		value[6:8],
		value[8:10],
		value[10:16],
	), nil
}
