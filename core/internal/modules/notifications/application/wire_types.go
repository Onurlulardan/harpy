package application

import (
	"io"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/google/uuid"
)

type ID = domain.ID
type Message = domain.Message
type Preference = domain.Preference
type Channel = domain.Channel

const (
	ChannelEmail    = domain.ChannelEmail
	ChannelTelegram = domain.ChannelTelegram
)

func ParseID(value string) (ID, error) { return domain.ParseID(value) }

func NewID(at time.Time, entropy io.Reader) (ID, error) {
	return domain.NewV7(at, entropy)
}

func IDFromUUID(value uuid.UUID) (ID, error) { return domain.IDFromBytes(value[:]) }
