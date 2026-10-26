package tenantsettings

import (
	"time"

	"github.com/google/uuid"
)

type RegistrySnapshot struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	Status    string
	Revision  int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SMTPView struct {
	Host               string
	Port               int
	Username           string
	From               string
	PasswordConfigured bool
}

type Subscription struct {
	PlanName            string
	PlanSlug            string
	Status              string
	MaxUsers            int
	MaxForms            int
	MaxStorageMB        int64
	SubmissionsPerMonth int64
}

type Projection struct {
	UserCount              int64
	FormCount              int64
	PendingInvitationCount int64
	SMTP                   *SMTPView
	Subscription           *Subscription
}

type Snapshot struct {
	RegistrySnapshot
	Projection
}

type SMTPMutation struct {
	Host, Username, From string
	Port                 int
	Password             []byte
}

type UpdateCommand struct {
	Name             *string
	SMTP             *SMTPMutation
	ExpectedRevision int64
	IdempotencyKey   string
}

type UpdateResult struct {
	Revision int64
	Replay   bool
}

type TestEmailCommand struct {
	To               string
	ExpectedRevision int64
	IdempotencyKey   string
}

type TestEmailResult struct {
	OperationID uuid.UUID
	Replay      bool
}
