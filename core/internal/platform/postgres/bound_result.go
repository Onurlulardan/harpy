package postgres

type CommandResult struct {
	rowsAffected int64
}

func (result CommandResult) RowsAffected() int64 { return result.rowsAffected }

type Row interface {
	Scan(destinations ...any) error
}

type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(destinations ...any) error
}
