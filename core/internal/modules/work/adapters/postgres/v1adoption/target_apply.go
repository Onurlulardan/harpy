package v1adoption

import (
	"context"
	"strconv"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

var issueNamespace = uuid.MustParse("3a238ae9-d55a-5484-91d0-e9cae9ba0c30")

func (store *TargetStore) PersistPlan(ctx context.Context, scope ports.Scope,
	plan adoption.Plan,
) (adoption.Receipt, error) {
	if err := store.check(ctx, scope); err != nil {
		return adoption.Receipt{}, err
	}
	if !plan.RunID.Valid() || plan.Generation < 1 || !plan.ActorID.Valid() ||
		plan.ManifestSHA256 == [32]byte{} || plan.SourceSHA256 == [32]byte{} ||
		plan.SourceCount < 0 || plan.CreatedAt.IsZero() {
		return adoption.Receipt{}, ports.ErrInvalid
	}
	targetHash := plan.TargetSHA256()
	if err := store.executeOne(ctx, beginAdoptionSQL, plan.RunID, plan.Generation,
		plan.ManifestSHA256[:], plan.SourceCount, len(plan.Records),
		plan.SourceSHA256[:], targetHash[:], plan.CreatedAt); err != nil {
		return adoption.Receipt{}, err
	}
	if len(plan.Issues) > 0 {
		return store.block(ctx, plan, targetHash)
	}
	if err := store.executeOne(ctx, fenceWriterSQL, plan.RunID, plan.Generation,
		plan.ActorID); err != nil {
		return adoption.Receipt{}, err
	}
	if err := store.executeOne(ctx, setAdoptionStateSQL, plan.RunID, "COPYING",
		[]string{"INVENTORY", "COPYING"}); err != nil {
		return adoption.Receipt{}, err
	}
	for _, record := range plan.Records {
		if err := store.writer.Persist(ctx, record); err != nil {
			return adoption.Receipt{}, err
		}
		if err := store.executeOne(ctx, recordObjectSQL, plan.RunID, record.SourceKind,
			record.SourceID, record.Kind, record.ID, record.SourceSHA256[:], record.SHA256[:]); err != nil {
			return adoption.Receipt{}, err
		}
	}
	if err := store.executeOne(ctx, setAdoptionStateSQL, plan.RunID, "RECONCILING",
		[]string{"COPYING", "RECONCILING"}); err != nil {
		return adoption.Receipt{}, err
	}
	return adoption.Receipt{RunID: plan.RunID, State: "RECONCILING",
		SourceCount: plan.SourceCount, TargetCount: int64(len(plan.Records)), SHA256: targetHash}, nil
}

func (store *TargetStore) block(ctx context.Context, plan adoption.Plan,
	targetHash [32]byte,
) (adoption.Receipt, error) {
	for _, issue := range plan.Issues {
		identity := string(issue.SourceKind) + ":" + issue.SourceID.String() + ":" + string(issue.Code)
		issueID := uuid.NewSHA1(issueNamespace, []byte(plan.RunID.String()+":"+identity))
		if err := store.executeOne(ctx, recordIssueSQL, plan.RunID, issueID,
			issue.SourceKind, issue.SourceID, issue.Code, issue.Evidence[:]); err != nil {
			return adoption.Receipt{}, err
		}
	}
	if err := store.executeOne(ctx, setAdoptionStateSQL, plan.RunID, "BLOCKED",
		[]string{"INVENTORY", "BLOCKED"}); err != nil {
		return adoption.Receipt{}, err
	}
	return adoption.Receipt{RunID: plan.RunID, State: "BLOCKED",
		SourceCount: plan.SourceCount, TargetCount: 0,
		SHA256: targetHash}, nil
}

func objectKey(kind adoption.TargetKind, id string) string {
	return string(kind) + ":" + id + ":" + strconv.Itoa(len(id))
}
