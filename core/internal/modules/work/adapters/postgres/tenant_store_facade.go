package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type row interface{ Scan(...any) error }
type rows interface {
	row
	Next() bool
	Err() error
	Close()
}
type commandResult struct{ affected int64 }

func (result commandResult) RowsAffected() int64 { return result.affected }

type databaseFacade struct {
	executeFn     func(context.Context, string, ...any) (int64, error)
	query         func(context.Context, string, ...any) (rows, error)
	queryRow      func(context.Context, string, ...any) row
	currentSchema func(context.Context) (string, error)
}

func (database databaseFacade) run(ctx context.Context, sql string, arguments ...any) (commandResult, error) {
	affected, err := database.executeFn(ctx, sql, arguments...)
	return commandResult{affected: affected}, err
}
func (database databaseFacade) Query(ctx context.Context, sql string, arguments ...any) (rows, error) {
	return database.query(ctx, sql, arguments...)
}
func (database databaseFacade) QueryRow(ctx context.Context, sql string, arguments ...any) row {
	return database.queryRow(ctx, sql, arguments...)
}
func (database databaseFacade) CurrentSchema(ctx context.Context) (string, error) {
	return database.currentSchema(ctx)
}
func (database databaseFacade) platformBindingAvailable(ctx context.Context) bool {
	if database.currentSchema == nil {
		return false
	}
	_, err := database.CurrentSchema(ctx)
	return err == nil
}

type Store struct {
	database databaseFacade
	bridge   *scopeBridge
}

func newStore(bridge *scopeBridge, database databaseFacade) *Store {
	return &Store{database: database, bridge: bridge}
}
func (store *Store) Boards() ports.BoardStore                { return store }
func (store *Store) Tasks() ports.TaskStore                  { return store }
func (store *Store) Collaboration() ports.CollaborationStore { return store }
func (store *Store) Calendar() ports.CalendarStore           { return store }
func (store *Store) Projections() ports.ProjectionStore      { return store }
func (store *Store) Operations() ports.OperationStore        { return store }

func (store *Store) check(ctx context.Context) error {
	if store == nil || !scopeUsable(ctx, store.bridge) {
		return application.ErrUnavailable
	}
	if !store.database.platformBindingAvailable(ctx) {
		return application.ErrUnavailable
	}
	return nil
}

func executeOne(ctx context.Context, store *Store, sql string, arguments ...any) error {
	if err := store.check(ctx); err != nil {
		return err
	}
	result, err := store.database.run(ctx, sql, arguments...)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return application.ErrConflict
	}
	return nil
}

var _ ports.Stores = (*Store)(nil)
var _ ports.BoardStore = (*Store)(nil)
var _ ports.TaskStore = (*Store)(nil)
var _ ports.CollaborationStore = (*Store)(nil)
var _ ports.CalendarStore = (*Store)(nil)
var _ ports.ProjectionStore = (*Store)(nil)
var _ ports.OperationStore = (*Store)(nil)
