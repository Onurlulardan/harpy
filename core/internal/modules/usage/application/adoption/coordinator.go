package adoption

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

type Coordinator struct {
	source Source
	target Target
	clock  Clock
}

type ExecuteRequest struct {
	OperationID, RunID, LeaseOwner uuid.UUID
	PageLimit, PageBudget          int
}

func NewCoordinator(source Source, target Target, clock Clock) (*Coordinator, error) {
	if nilPort(source) || nilPort(target) || nilPort(clock) {
		return nil, fmt.Errorf("usage adoption dependencies are invalid")
	}
	return &Coordinator{source: source, target: target, clock: clock}, nil
}

func (coordinator *Coordinator) Execute(ctx context.Context, request ExecuteRequest) (Run, error) {
	if coordinator == nil || ctx == nil || ctx.Err() != nil || request.OperationID == uuid.Nil ||
		request.RunID == uuid.Nil || request.RunID.Version() != 7 || request.LeaseOwner == uuid.Nil ||
		request.PageLimit < 1 || request.PageLimit > 100 ||
		request.PageBudget < 1 || request.PageBudget > 10000 {
		return Run{}, domain.ErrInvalid
	}
	manifest, err := coordinator.source.Inspect(ctx)
	if err != nil || manifest.Validate() != nil {
		return Run{}, domain.ErrUnavailable
	}
	now := coordinator.clock.Now().UTC()
	run, _, err := coordinator.target.Start(ctx, StartCommand{OperationID: request.OperationID,
		RunID: request.RunID, Manifest: manifest, TransformSHA256: TransformSHA256,
		ObservedAt: now})
	if err != nil {
		return Run{}, err
	}
	if err = validateRun(run, request, manifest); err != nil || run.Status == StatusCompleted {
		return run, err
	}
	run, err = coordinator.claim(ctx, run, request, manifest)
	if err != nil {
		return run, err
	}
	if manifest.Kind == ZeroSource {
		return coordinator.finish(ctx, run, request, manifest)
	}
	for pages := 0; pages < request.PageBudget && run.CopiedCount < manifest.RowCount; pages++ {
		if err = coordinator.checkSource(ctx, run, request, manifest); err != nil {
			return run, err
		}
		page, pageErr := coordinator.source.Page(ctx, manifest.SnapshotID,
			run.CursorID, request.PageLimit)
		if pageErr != nil || page.Validate(request.PageLimit) != nil ||
			run.CopiedCount+int64(len(page.Records)) > manifest.RowCount {
			return run, domain.ErrSourceDrift
		}
		records, targetSHA, transformErr := TransformPage(manifest, page)
		if transformErr != nil {
			return run, transformErr
		}
		command := ApplyPageCommand{LeaseCommand: leaseCommand(run, request, manifest,
			coordinator.clock.Now().UTC()), PageNumber: run.NextPage,
			CursorBefore: run.CursorID, CursorAfter: page.CursorAfter, Records: records,
			SourcePageSHA256: page.SHA256, TargetSHA256: targetSHA}
		if _, err = coordinator.target.ApplyPage(ctx, command); err != nil {
			return run, err
		}
		run, err = coordinator.target.Advance(ctx, command)
		if err != nil || run.Validate() != nil {
			return run, errOrUnavailable(err)
		}
	}
	if run.CopiedCount == manifest.RowCount {
		return coordinator.finish(ctx, run, request, manifest)
	}
	return run, nil
}

func (coordinator *Coordinator) claim(ctx context.Context, run Run,
	request ExecuteRequest, manifest Manifest,
) (Run, error) {
	current, err := coordinator.source.Inspect(ctx)
	if err != nil || current.Validate() != nil {
		return run, domain.ErrUnavailable
	}
	claimed, err := coordinator.target.Claim(ctx, ClaimCommand{OperationID: request.OperationID,
		RunID: request.RunID, LeaseOwner: request.LeaseOwner, ExpectedRevision: run.Revision,
		Manifest: current, ObservedAt: coordinator.clock.Now().UTC()})
	if err != nil {
		return run, err
	}
	if claimed.Status == StatusFenced || !current.Same(manifest) {
		return claimed, domain.ErrSourceDrift
	}
	if claimed.Validate() != nil {
		return Run{}, domain.ErrUnavailable
	}
	return claimed, nil
}
