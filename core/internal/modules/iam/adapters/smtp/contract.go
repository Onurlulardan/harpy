package smtp

import (
	"errors"
	"time"
)

var ErrContract = errors.New("IAM invitation SMTP contract is invalid")

type ResultKind uint8

const (
	ResultInvalid ResultKind = iota
	ResultAccepted
	ResultTransient
	ResultPermanent
	ResultUnknown
)

type Result struct {
	Kind              ResultKind
	Code              string
	ProviderReference string
	EvidenceSHA256    [32]byte
	CompletedAt       time.Time
}

type Message struct {
	Recipient  string
	MessageID  string
	TenantSlug string
	Token      []byte
	Date       time.Time
}
