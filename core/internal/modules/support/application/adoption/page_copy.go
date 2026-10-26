package adoption

import (
	"context"
	"fmt"
	"time"
)

func (worker *Worker) copyPage(ctx context.Context, locator VerifiedLocator, run Run,
	observedAt time.Time,
) (Run, bool, error) {
	var records []Record
	var done bool
	if run.Kind == ContactGlobal {
		page, err := worker.source.ReadContacts(ctx, locator, run.CursorID, PageLimit)
		if err != nil {
			return Run{}, false, ErrSourceDrift
		}
		records, err = contactRecords(run, page)
		done = page.Done
		if err != nil {
			return Run{}, false, err
		}
	} else {
		page, err := worker.source.ReadBugs(ctx, locator, run.CursorID, PageLimit)
		if err != nil {
			return Run{}, false, ErrSourceDrift
		}
		records, err = bugRecords(run, page)
		done = page.Done
		if err != nil {
			return Run{}, false, err
		}
	}
	if len(records) == 0 {
		return run, done, nil
	}
	command := ApplyPageCommand{OperationID: run.OperationID, RunID: run.RunID,
		LeaseOwner: run.LeaseOwner, Kind: run.Kind, LeaseGeneration: run.LeaseGeneration,
		PageNumber: run.NextPage, CursorBefore: run.CursorID, Records: records,
		SourcePageSHA256: PageSHA256(records), ObservedAt: observedAt}
	receipt, err := worker.target.AdoptionExecutePage(ctx, command)
	if err != nil {
		return Run{}, false, err
	}
	if receipt.RowCount != int64(len(records)) ||
		receipt.SourcePageSHA256 != command.SourcePageSHA256 ||
		receipt.CursorAfter != records[len(records)-1].SourceID {
		return Run{}, false, fmt.Errorf("%w: page receipt mismatch", ErrUnavailable)
	}
	run, err = worker.target.AdoptionRecordAdvance(ctx, command)
	if err != nil {
		return Run{}, false, err
	}
	if !run.valid() {
		return Run{}, false, fmt.Errorf("%w: advanced run document is invalid", ErrUnavailable)
	}
	return run, done, nil
}

func normalizeTargetResult(err error) error {
	if err != nil {
		return err
	}
	return ErrUnavailable
}
