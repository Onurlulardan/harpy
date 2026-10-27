package v1adoption

import (
	"context"
	_ "embed"
	"regexp"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

//go:embed sql/tenant/read_task_boards.sql
var readBoardsSQL string

//go:embed sql/tenant/read_task_board_columns.sql
var readColumnsSQL string

//go:embed sql/tenant/read_board_members.sql
var readMembershipsSQL string

//go:embed sql/tenant/read_board_tasks.sql
var readTasksSQL string

//go:embed sql/tenant/read_board_task_comments.sql
var readCommentsSQL string

//go:embed sql/tenant/read_calendar_events.sql
var readCalendarSQL string

var schemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type Rows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}

type DatabaseAccess struct {
	Execute       func(context.Context, string, ...any) (int64, error)
	Query         func(context.Context, string, ...any) (Rows, error)
	QueryRow      func(context.Context, string, ...any) Row
	CurrentSchema func(context.Context) (string, error)
}

type Row interface{ Scan(...any) error }

type LegacySource struct {
	database DatabaseAccess
	tenantID domain.ID
}

func NewLegacySource(database DatabaseAccess, tenantID domain.ID) (*LegacySource, error) {
	if database.Query == nil || database.CurrentSchema == nil || !tenantID.Valid() {
		return nil, ports.ErrInvalid
	}
	return &LegacySource{database: database, tenantID: tenantID}, nil
}

func (source *LegacySource) LoadLegacy(ctx context.Context, scope ports.Scope) ([]adoption.SourceRecord, error) {
	if source == nil || scope == nil || scope.TenantID() != source.tenantID ||
		ctx == nil || ctx.Err() != nil {
		return nil, ports.ErrInvalid
	}
	schema, err := source.database.CurrentSchema(ctx)
	if err != nil || !schemaPattern.MatchString(schema) || schema == "public" {
		return nil, ports.ErrUnavailable
	}
	result := []adoption.SourceRecord{}
	for _, read := range []func(context.Context) ([]adoption.SourceRecord, error){
		source.readBoards, source.readColumns, source.readMemberships,
		source.readTasks, source.readComments, source.readCalendar,
	} {
		records, readErr := read(ctx)
		if readErr != nil {
			return nil, readErr
		}
		result = append(result, records...)
	}
	return result, nil
}

var _ adoption.Source = (*LegacySource)(nil)
