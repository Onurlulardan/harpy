package postgres

import (
	"context"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	"github.com/google/uuid"
)

type row interface{ Scan(...any) error }
type rows interface {
	Close()
	Err() error
	Next() bool
	Scan(...any) error
}
type commandResult interface{ RowsAffected() int64 }
type queryRow func(context.Context, string, ...any) row
type queryMany func(context.Context, string, ...any) (rows, error)
type execute func(context.Context, string, ...any) (commandResult, error)
type scopeCheck func(context.Context) bool
type rowNotFound func(error) bool

type EnqueueJob func(context.Context, jobsv1.Enqueue) (string, error)

type authorityStamp struct {
	domain, groupID      string
	generation, groupGen int64
}

type Store struct {
	queryRow  queryRow
	query     queryMany
	execute   execute
	allowed   scopeCheck
	notFound  rowNotFound
	command   bool
	tenantID  uuid.UUID
	authority authorityStamp
	jobs      jobsv1.Catalog
	enqueue   EnqueueJob
}

func (store *Store) usable(ctx context.Context, command bool) bool {
	return store != nil && store.queryRow != nil && store.query != nil &&
		store.execute != nil && store.allowed != nil && store.notFound != nil &&
		ctx != nil && ctx.Err() == nil && store.allowed(ctx) && (!command || store.command)
}

func (store *Store) Effects() ports.EffectStore            { return store }
func (store *Store) Messages() ports.MessageStore          { return store }
func (store *Store) Preferences() ports.PreferenceStore    { return store }
func (store *Store) Templates() ports.TemplateStore        { return store }
func (store *Store) Providers() ports.ProviderConfigStore  { return store }
func (store *Store) SMTPSettings() ports.SMTPSettingsStore { return store }
func (store *Store) Deliveries() ports.DeliveryStore       { return store }
func (store *Store) Realtime() ports.RealtimeStore         { return store }
func (store *Store) Evidence() ports.EvidenceStore         { return store }

var _ ports.Stores = (*Store)(nil)
