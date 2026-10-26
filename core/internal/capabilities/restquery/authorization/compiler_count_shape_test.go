package authorization

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type compilerCountBatchConsumer struct {
	record compiler.RecordStatement
	count  compiler.ExactCountStatement
}

func (consumer *compilerCountBatchConsumer) ConsumeRecord(
	statement compiler.RecordStatement,
) error {
	consumer.record = statement
	return nil
}

func (consumer *compilerCountBatchConsumer) ConsumeExactCount(
	statement compiler.ExactCountStatement,
) error {
	consumer.count = statement
	return nil
}

func TestCompilerExactCountAndBatchRejectUnsupportedShapes(t *testing.T) {
	root := snapshotRecordSpec(t, 1)
	countNone := compilerBoundaryPlan(t, root)
	includeRoot := snapshotRecordSpec(t, 2)
	withoutInclude := compilerBoundaryPlan(
		t, compilerBoundaryExactSpec(t, includeRoot),
	)
	include := compilerBoundaryPlan(
		t, compilerBoundaryExactIncludeSpec(t, includeRoot),
	)
	aggregate := compilerBoundaryPlan(t,
		authorizationAggregateSpec(t, newAuthorizationQueryFixture(t)))
	postgresCompiler, err := compiler.New(compilerBoundaryCodec(t))
	if err != nil {
		t.Fatal(err)
	}
	if authorized, proofErr := countNone.ExactCountAuthorized(); proofErr != nil || authorized {
		t.Fatal("count=none plan claimed exact-count authorization")
	}
	if _, err := postgresCompiler.CompileExactCount(countNone); !errors.Is(err, compiler.ErrContract) {
		t.Fatal("count=none minted exact-count statement")
	}
	assertCompilerCountIncludeParity(t, postgresCompiler, withoutInclude, include)
	if _, err := postgresCompiler.CompileExactCount(aggregate); !errors.Is(err, compiler.ErrContract) {
		t.Fatal("aggregate minted exact-count statement")
	}
	if batch, err := postgresCompiler.CompileRecordBatch(countNone); err != nil ||
		batch.ValidatePlan(countNone) != nil {
		t.Fatalf("count=none batch failed: %v", err)
	}
	if _, err := postgresCompiler.CompileRecordBatch(aggregate); !errors.Is(err, compiler.ErrContract) {
		t.Fatal("aggregate minted root record batch")
	}
}

func assertCompilerCountIncludeParity(
	t *testing.T,
	postgresCompiler compiler.Compiler,
	without, with resolved.AuthorizedPlan,
) {
	t.Helper()
	withoutCount, err := postgresCompiler.CompileExactCount(without)
	if err != nil {
		t.Fatal(err)
	}
	withCount, err := postgresCompiler.CompileExactCount(with)
	if err != nil {
		t.Fatal(err)
	}
	assertCompilerStatementParity(t, withoutCount, withCount)
	withoutBatch, err := postgresCompiler.CompileRecordBatch(without)
	if err != nil {
		t.Fatal(err)
	}
	withBatch, err := postgresCompiler.CompileRecordBatch(with)
	if err != nil {
		t.Fatal(err)
	}
	withoutConsumer := &compilerCountBatchConsumer{}
	withConsumer := &compilerCountBatchConsumer{}
	if compiler.ConsumeRecordBatch(withoutBatch, without, withoutConsumer) != nil ||
		compiler.ConsumeRecordBatch(withBatch, with, withConsumer) != nil {
		t.Fatal("root count batch consumption failed")
	}
	assertCompilerStatementParity(t, withoutConsumer.record, withConsumer.record)
	assertCompilerStatementParity(t, withoutConsumer.count, withConsumer.count)
}

type compilerSQLArguments interface {
	SQL() (string, error)
	Arguments() ([]any, error)
}

func assertCompilerStatementParity(
	t *testing.T, without, with compilerSQLArguments,
) {
	t.Helper()
	withoutSQL, withoutSQLErr := without.SQL()
	withSQL, withSQLErr := with.SQL()
	withoutArguments, withoutArgumentsErr := without.Arguments()
	withArguments, withArgumentsErr := with.Arguments()
	if withoutSQLErr != nil || withSQLErr != nil || withoutArgumentsErr != nil ||
		withArgumentsErr != nil || withoutSQL != withSQL ||
		!reflect.DeepEqual(withoutArguments, withArguments) {
		t.Fatal("include state changed root record/count SQL semantics")
	}
}
