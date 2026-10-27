package authorization

import (
	"errors"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type compilerBatchCollector struct {
	record compiler.RecordStatement
	count  compiler.ExactCountStatement
	roles  []planner.StatementClass
}

func (collector *compilerBatchCollector) ConsumeRecord(
	statement compiler.RecordStatement,
) error {
	collector.record = statement
	collector.roles = append(collector.roles, planner.StatementClassRecord)
	return nil
}

func (collector *compilerBatchCollector) ConsumeExactCount(
	statement compiler.ExactCountStatement,
) error {
	collector.count = statement
	collector.roles = append(collector.roles, planner.StatementClassExactCount)
	return nil
}

func TestCompilerExactCountAndAtomicBatchBoundary(t *testing.T) {
	root := snapshotRecordSpec(t, 1)
	exactSpec := compilerBoundaryExactSpec(t, root)
	plan := compilerBoundaryPlan(t, exactSpec)
	codec := compilerBoundaryCodec(t)
	postgresCompiler, err := compiler.New(codec)
	if err != nil {
		t.Fatal(err)
	}
	authorized, proofErr := plan.ExactCountAuthorized()
	if proofErr != nil || !authorized {
		t.Fatal("authorized exact-count proof was missing")
	}
	count, countErr := postgresCompiler.CompileExactCount(plan)
	batch, batchErr := postgresCompiler.CompileRecordBatch(plan)
	collector := &compilerBatchCollector{}
	consumeErr := compiler.ConsumeRecordBatch(batch, plan, collector)
	mode, modeErr := batch.CountMode()
	if countErr != nil || batchErr != nil || consumeErr != nil || modeErr != nil ||
		mode != api.CountModeExact || len(collector.roles) != 2 ||
		batch.ValidatePlan(plan) != nil || count.ValidatePlan(plan) != nil ||
		collector.record.ValidatePlan(plan) != nil || collector.count.ValidatePlan(plan) != nil {
		t.Fatalf("exact boundary count=%v batch=%v consume=%v", countErr, batchErr, consumeErr)
	}
	countClass, _ := count.Class()
	recordClass, _ := collector.record.Class()
	recordCountMode, _ := collector.record.CountMode()
	countSQL, _ := count.SQL()
	recordSQL, _ := collector.record.SQL()
	if countClass != planner.StatementClassExactCount ||
		recordClass != planner.StatementClassRecord || recordCountMode != api.CountModeExact ||
		!strings.HasPrefix(countSQL, "SELECT COUNT(*) FROM ") ||
		strings.Contains(countSQL, "ORDER BY") || strings.Contains(countSQL, "LIMIT") ||
		!strings.Contains(recordSQL, "ORDER BY") || !strings.Contains(recordSQL, "LIMIT") {
		t.Fatalf("count SQL=%s record SQL=%s", countSQL, recordSQL)
	}
	assertCompilerRejectsCrossPlan(t, collector.record, plan)
	assertCompilerRejectsCrossPlan(t, count, plan)
	assertCompilerRejectsCrossPlan(t, batch, plan)
}

func TestCompilerBatchBindsExactContinuationToken(t *testing.T) {
	root := snapshotRecordSpec(t, 1)
	exactSpec := compilerBoundaryExactSpec(t, root)
	exactPlan := compilerBoundaryPlan(t, exactSpec)
	codec := compilerBoundaryCodec(t)
	postgresCompiler, _ := compiler.New(codec)
	baseBatch, err := postgresCompiler.CompileRecordBatch(exactPlan)
	base := &compilerBatchCollector{}
	if err != nil || compiler.ConsumeRecordBatch(baseBatch, exactPlan, base) != nil {
		t.Fatal("could not collect exact root cursor context")
	}
	tokenA := compilerBoundaryCursor(t, codec, base.record)
	tokenB := compilerBoundaryCursor(t, codec, base.record)
	if tokenA == tokenB {
		t.Fatal("cursor codec reused an encrypted continuation token")
	}
	planA := compilerBoundaryPlan(
		t, compilerBoundaryExactAfterSpec(t, root, tokenA),
	)
	planB := compilerBoundaryPlan(
		t, compilerBoundaryExactAfterSpec(t, root, tokenB),
	)
	batchA, err := postgresCompiler.CompileRecordBatch(planA)
	collectedA := &compilerBatchCollector{}
	if err != nil || compiler.ConsumeRecordBatch(batchA, planA, collectedA) != nil {
		t.Fatalf("valid exact continuation did not compile: %v", err)
	}
	standaloneA, err := postgresCompiler.CompileExactCount(planA)
	if err != nil {
		t.Fatalf("valid exact-count continuation failed: %v", err)
	}
	afterCountSQL, _ := standaloneA.SQL()
	if strings.Contains(afterCountSQL, " WHERE ") ||
		strings.Contains(afterCountSQL, "ORDER BY") ||
		strings.Contains(afterCountSQL, "LIMIT") {
		t.Fatalf("continuation changed the counted root set: %s", afterCountSQL)
	}
	assertContinuationTransplantRejected(
		t, planB, batchA, collectedA.record, collectedA.count, standaloneA,
	)
	rejected := &compilerBatchCollector{}
	if !errors.Is(
		compiler.ConsumeRecordBatch(batchA, planB, rejected), compiler.ErrContract,
	) || len(rejected.roles) != 0 {
		t.Fatal("cross-cursor batch exposed leaves before plan validation")
	}
	invalidPlan := compilerBoundaryPlan(
		t, compilerBoundaryExactAfterSpec(t, root, "raw-invalid-cursor"),
	)
	if _, err := postgresCompiler.CompileRecordBatch(invalidPlan); err == nil {
		t.Fatal("invalid cursor minted a record batch")
	}
	if _, err := postgresCompiler.CompileExactCount(invalidPlan); err == nil {
		t.Fatal("invalid cursor minted an exact-count statement")
	}
	assertContinuationTransplantRejected(
		t, invalidPlan, batchA, collectedA.record, collectedA.count, standaloneA,
	)
	rejected = &compilerBatchCollector{}
	if !errors.Is(
		compiler.ConsumeRecordBatch(batchA, invalidPlan, rejected), compiler.ErrContract,
	) || len(rejected.roles) != 0 {
		t.Fatal("invalid-cursor plan exposed batch leaves")
	}
}

func assertContinuationTransplantRejected(
	t *testing.T,
	plan resolved.AuthorizedPlan,
	values ...compilerPlanValidator,
) {
	t.Helper()
	for _, value := range values {
		if !errors.Is(value.ValidatePlan(plan), compiler.ErrContract) {
			t.Fatal("continuation token transplant was accepted")
		}
	}
}
