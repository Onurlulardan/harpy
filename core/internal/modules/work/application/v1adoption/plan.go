package v1adoption

import (
	"bytes"
	"context"
	"crypto/sha256"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) buildPlan(ctx context.Context, scope ports.Scope,
	request ApplyRequest, raw []SourceRecord,
) (Plan, error) {
	records, err := validateSource(raw)
	if err != nil {
		return Plan{}, err
	}
	index, err := indexLegacy(records)
	if err != nil {
		return Plan{}, err
	}
	plan := Plan{RunID: request.RunID, ActorID: scope.ActorID(),
		Generation: request.Generation, ManifestSHA256: request.ManifestSHA256,
		SourceSHA256: sourceDigest(records), SourceCount: int64(len(records)),
		CreatedAt: service.dependencies.Clock.Now().UTC()}
	for _, transform := range []func(*Plan, legacyIndex) error{
		transformBoards, transformMemberships,
	} {
		if err = transform(&plan, index); err != nil {
			return Plan{}, err
		}
	}
	if err = service.transformTasks(ctx, scope, &plan, index); err != nil {
		return Plan{}, err
	}
	if err = transformComments(&plan, index); err != nil {
		return Plan{}, err
	}
	if err = service.transformCalendar(ctx, scope, &plan, index); err != nil {
		return Plan{}, err
	}
	sortPlan(&plan)
	return plan, nil
}

func sortPlan(plan *Plan) {
	sort.Slice(plan.Records, func(left, right int) bool {
		if plan.Records[left].Kind != plan.Records[right].Kind {
			return plan.Records[left].Kind < plan.Records[right].Kind
		}
		return bytes.Compare(plan.Records[left].ID[:], plan.Records[right].ID[:]) < 0
	})
	sort.Slice(plan.Issues, func(left, right int) bool {
		if plan.Issues[left].SourceKind != plan.Issues[right].SourceKind {
			return plan.Issues[left].SourceKind < plan.Issues[right].SourceKind
		}
		if compared := bytes.Compare(plan.Issues[left].SourceID[:], plan.Issues[right].SourceID[:]); compared != 0 {
			return compared < 0
		}
		return plan.Issues[left].Code < plan.Issues[right].Code
	})
}

func targetDigest(records []TargetRecord) [sha256.Size]byte {
	hash := sha256.New()
	for _, record := range records {
		_, _ = hash.Write([]byte(record.Kind))
		_, _ = hash.Write(record.ID[:])
		_, _ = hash.Write(record.SHA256[:])
	}
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func (plan Plan) TargetSHA256() [sha256.Size]byte { return targetDigest(plan.Records) }
