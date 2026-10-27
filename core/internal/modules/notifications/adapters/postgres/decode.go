package postgres

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func decodeID(value string) (domain.ID, error) {
	result, err := domain.ParseID(value)
	if err != nil {
		return domain.ID{}, fmt.Errorf("decode notification identity")
	}
	return result, nil
}

func decodeDigest(value []byte) ([32]byte, error) {
	var result [32]byte
	if len(value) != len(result) {
		return result, fmt.Errorf("decode notification digest")
	}
	copy(result[:], value)
	return result, nil
}
