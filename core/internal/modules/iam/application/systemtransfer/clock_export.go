package systemtransfer

import "time"

type ExportClock interface {
	Now() time.Time
}
