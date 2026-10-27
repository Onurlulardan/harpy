package v1bridge

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type Writer string

const (
	WriterV1 Writer = "v1"
	WriterV2 Writer = "v2"
)

type Authority struct {
	ProducerWriter, NotificationWriter Writer
	ProducerState, NotificationState   string
	ProducerGeneration                 int64
	NotificationGeneration             int64
}

func (value Authority) valid() bool {
	if value.ProducerWriter != WriterV2 || value.ProducerState != "v2" ||
		value.ProducerGeneration < 1 || value.NotificationGeneration < 1 {
		return false
	}
	return value.NotificationWriter == WriterV1 &&
		(value.NotificationState == "v1" || value.NotificationState == "shadow") ||
		value.NotificationWriter == WriterV2 && value.NotificationState == "v2"
}

type State string

const (
	StateQueued         State = "queued"
	StateAccepted       State = "accepted"
	StateFailed         State = "failed"
	StateOutcomeUnknown State = "outcome_unknown"
)

type Intent struct {
	Producer                                      string
	EffectKeyHash, CanonicalDigest, PayloadDigest [32]byte
	RecipientID, JobID                            domain.ID
	ProducerGeneration, NotificationGeneration    int64
	Payload                                       []byte
	State                                         State
	CreatedAt, UpdatedAt                          time.Time
	Revision                                      int64
}

type Job struct {
	ID, TenantID                               domain.ID
	Producer                                   string
	ProducerGeneration, NotificationGeneration int64
	PayloadReference, PayloadDigestHex         string
	EffectKey                                  string
	RunAt                                      time.Time
}
