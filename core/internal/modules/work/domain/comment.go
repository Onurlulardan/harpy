package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrCommentRedaction = errors.New("work comment redaction is invalid")

type RedactionKind string

const (
	RedactionAuthor    RedactionKind = "AUTHOR"
	RedactionModerator RedactionKind = "MODERATOR"
)

type Comment struct {
	ID, BoardID, TaskID, AuthorID ID
	Body                          string
	CreatedAt                     time.Time
	RedactedAt                    *time.Time
	RedactedBy                    ID
	Redaction                     RedactionKind
	RedactionReason               string
	RowVersion                    int64
}

func (comment Comment) Valid() bool {
	if !comment.ID.Valid() || !comment.BoardID.Valid() || !comment.TaskID.Valid() ||
		!comment.AuthorID.Valid() || comment.CreatedAt.IsZero() || comment.RowVersion < 1 {
		return false
	}
	if comment.RedactedAt == nil {
		return strings.TrimSpace(comment.Body) != "" && len(comment.Body) <= 16_000 &&
			!comment.RedactedBy.Valid() && comment.Redaction == "" && comment.RedactionReason == ""
	}
	return comment.Body == "" && comment.RedactedBy.Valid() &&
		(comment.Redaction == RedactionAuthor && comment.RedactionReason == "" ||
			comment.Redaction == RedactionModerator && strings.TrimSpace(comment.RedactionReason) != "" &&
				len(comment.RedactionReason) <= 1000)
}

func (comment Comment) Redact(actor ID, moderator bool, reason string, at time.Time) (Comment, error) {
	if !comment.Valid() || comment.RedactedAt != nil || !actor.Valid() || at.IsZero() || at.Before(comment.CreatedAt) {
		return Comment{}, ErrCommentRedaction
	}
	if actor == comment.AuthorID {
		comment.Redaction = RedactionAuthor
	} else if moderator && strings.TrimSpace(reason) != "" && len(reason) <= 1000 {
		comment.Redaction = RedactionModerator
	} else {
		return Comment{}, ErrCommentRedaction
	}
	comment.Body = ""
	comment.RedactedAt = &at
	comment.RedactedBy = actor
	comment.RedactionReason = strings.TrimSpace(reason)
	comment.RowVersion++
	return comment, nil
}
