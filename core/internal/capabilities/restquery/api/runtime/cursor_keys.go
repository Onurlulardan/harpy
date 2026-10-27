package runtime

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

const cursorRedacted = "[REDACTED]"

type CursorKeys struct {
	currentID, previousID string
	current, previous     [32]byte
}

func NewCursorKeys(
	currentID string, current [32]byte,
	previousID string, previous [32]byte,
) (CursorKeys, error) {
	currentKey, err := config.NewCursorKey(currentID, current)
	if err != nil {
		return CursorKeys{}, err
	}
	previousKey, err := config.NewCursorKey(previousID, previous)
	if err != nil {
		return CursorKeys{}, err
	}
	if _, err := config.NewCursorKeyring(currentKey, previousKey); err != nil {
		return CursorKeys{}, err
	}
	return CursorKeys{
		currentID: currentID, current: current,
		previousID: previousID, previous: previous,
	}, nil
}

func (keys CursorKeys) keyring() (config.CursorKeyring, error) {
	current, err := config.NewCursorKey(keys.currentID, keys.current)
	if err != nil {
		return config.CursorKeyring{}, err
	}
	previous, err := config.NewCursorKey(keys.previousID, keys.previous)
	if err != nil {
		return config.CursorKeyring{}, err
	}
	return config.NewCursorKeyring(current, previous)
}

func (CursorKeys) String() string   { return cursorRedacted }
func (CursorKeys) GoString() string { return cursorRedacted }
func (CursorKeys) LogValue() slog.Value {
	return slog.StringValue(cursorRedacted)
}
func (CursorKeys) MarshalJSON() ([]byte, error) { return []byte(`"[REDACTED]"`), nil }
func (CursorKeys) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, cursorRedacted)
}
