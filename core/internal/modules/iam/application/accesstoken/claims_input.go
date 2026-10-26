package accesstoken

import (
	"time"

	"github.com/google/uuid"
)

type ClaimsInput struct {
	Subject                    uuid.UUID
	Session                    uuid.UUID
	Tenant                     uuid.UUID
	Client                     ClientID
	IssuedAt                   time.Time
	JTI                        uuid.UUID
	TokenVersion               int64
	AuthorizationVersion       int64
	PlatformAuthorizationClock int64
	SecurityEpoch              [32]byte
	AccessMode                 AccessMode
	PortalBinding              *uuid.UUID
	AuthenticationTime         time.Time
	Assurance                  Assurance
	Methods                    []AuthMethod
	StepUpPurposeHash          []byte
}
