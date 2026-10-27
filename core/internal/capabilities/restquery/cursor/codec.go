package cursor

import (
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type Codec struct {
	state *codecState
}

type codecState struct {
	current     codecKey
	previous    codecKey
	nonceSource io.Reader
	clock       func() time.Time
}

func NewCodec(keyring config.CursorKeyring) (Codec, error) {
	return newCodec(keyring, rand.Reader, time.Now)
}

func newCodec(
	keyring config.CursorKeyring,
	nonceSource io.Reader,
	clock func() time.Time,
) (Codec, error) {
	if keyring.Validate() != nil || nonceSource == nil || clock == nil {
		return Codec{}, errInvalidCodec
	}
	currentConfig, currentErr := keyring.Current()
	previousConfig, previousErr := keyring.Previous()
	current, currentKeyErr := newCodecKey(currentConfig)
	previous, previousKeyErr := newCodecKey(previousConfig)
	if currentErr != nil || previousErr != nil || currentKeyErr != nil || previousKeyErr != nil ||
		current.kid == previous.kid {
		return Codec{}, errInvalidCodec
	}
	return Codec{state: &codecState{
		current:     current,
		previous:    previous,
		nonceSource: nonceSource,
		clock:       clock,
	}}, nil
}

func (codec Codec) Validate() error {
	if codec.state == nil || codec.state.nonceSource == nil || codec.state.clock == nil ||
		codec.state.current.validate() != nil || codec.state.previous.validate() != nil ||
		codec.state.current.kid == codec.state.previous.kid {
		return errInvalidCodec
	}
	return nil
}

func (Codec) String() string                       { return redacted }
func (Codec) GoString() string                     { return redacted }
func (Codec) LogValue() slog.Value                 { return redactedLogValue() }
func (codec Codec) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Codec) MarshalJSON() ([]byte, error)         { return rejectJSONMarshal() }
