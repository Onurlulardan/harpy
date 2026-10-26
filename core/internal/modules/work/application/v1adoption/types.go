package v1adoption

import (
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type SourceKind string

const (
	KindBoard      SourceKind = "task_boards"
	KindColumn     SourceKind = "task_board_columns"
	KindMembership SourceKind = "board_members"
	KindTask       SourceKind = "board_tasks"
	KindComment    SourceKind = "board_task_comments"
	KindCalendar   SourceKind = "calendar_events"
)

func (kind SourceKind) Valid() bool {
	return kind == KindBoard || kind == KindColumn || kind == KindMembership ||
		kind == KindTask || kind == KindComment || kind == KindCalendar
}

type SourceRecord struct {
	Kind     SourceKind
	ID       domain.ID
	Document json.RawMessage
	SHA256   [sha256.Size]byte
}

type TargetKind string

const (
	TargetBoard      TargetKind = "BOARD"
	TargetColumn     TargetKind = "COLUMN"
	TargetMembership TargetKind = "MEMBERSHIP"
	TargetTask       TargetKind = "TASK"
	TargetAssignee   TargetKind = "ASSIGNEE"
	TargetLabel      TargetKind = "LABEL"
	TargetTaskLabel  TargetKind = "TASK_LABEL"
	TargetChecklist  TargetKind = "CHECKLIST"
	TargetComment    TargetKind = "COMMENT"
	TargetCalendar   TargetKind = "CALENDAR"
	TargetRecurrence TargetKind = "RECURRENCE"
	TargetFormLink   TargetKind = "FORM_LINK"
	TargetProjection TargetKind = "WORKFLOW_PROJECTION"
)

type TargetRecord struct {
	SourceKind   SourceKind
	SourceID     domain.ID
	Kind         TargetKind
	ID           domain.ID
	Document     json.RawMessage
	SourceSHA256 [sha256.Size]byte
	SHA256       [sha256.Size]byte
}

type IssueCode string

const (
	IssueInvalidEnum    IssueCode = "INVALID_ENUM"
	IssueCrossBoard     IssueCode = "CROSS_BOARD_REFERENCE"
	IssueOrphan         IssueCode = "ORPHAN"
	IssueUnresolvedForm IssueCode = "UNRESOLVED_FORM"
	IssueRecurrence     IssueCode = "INVALID_RECURRENCE"
	IssueChecksum       IssueCode = "CHECKSUM_MISMATCH"
	IssueFutureWork     IssueCode = "FUTURE_WORK_PENDING"
)

type Issue struct {
	Code       IssueCode
	SourceKind SourceKind
	SourceID   domain.ID
	Evidence   [sha256.Size]byte
}

type Plan struct {
	RunID, ActorID               domain.ID
	Generation                   int64
	ManifestSHA256, SourceSHA256 [sha256.Size]byte
	SourceCount                  int64
	Records                      []TargetRecord
	Issues                       []Issue
	CreatedAt                    time.Time
}

type Evidence struct {
	RunID                           domain.ID
	SourceCount, TargetCount        int64
	SourceSHA256, TargetSHA256      [sha256.Size]byte
	Accepted, WriterFenced          bool
	V2WriteCount, CompatibilityHits int64
}

type Receipt struct {
	RunID                    domain.ID
	State                    string
	SourceCount, TargetCount int64
	SHA256                   [sha256.Size]byte
}
