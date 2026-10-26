package systemtransfer

import "time"

func (worker *ExportWorker) exportNow() (time.Time, error) {
	if worker == nil || nilDependency(worker.clock) {
		return time.Time{}, ErrExportUnavailable
	}
	now := worker.clock.Now()
	if !exportArtifactTime(now) {
		return time.Time{}, ErrExportUnavailable
	}
	return now, nil
}
