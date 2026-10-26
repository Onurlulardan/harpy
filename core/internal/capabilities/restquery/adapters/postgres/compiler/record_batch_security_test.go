package compiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

var errBatchConsumerFixture = errors.New("batch consumer fixture failed")

type batchConsumerFixture struct {
	roles     []planner.StatementClass
	record    RecordStatement
	count     ExactCountStatement
	failCount bool
}

func (consumer *batchConsumerFixture) ConsumeRecord(statement RecordStatement) error {
	consumer.roles = append(consumer.roles, planner.StatementClassRecord)
	consumer.record = statement
	return nil
}

func (consumer *batchConsumerFixture) ConsumeExactCount(
	statement ExactCountStatement,
) error {
	consumer.roles = append(consumer.roles, planner.StatementClassExactCount)
	consumer.count = statement
	if consumer.failCount {
		return errBatchConsumerFixture
	}
	return nil
}

func TestRecordBatchConsumptionIsOrderedAndExact(t *testing.T) {
	input, _ := fixtureExactCountScenario(t)
	record, _ := compileRecord(input)
	count, _ := compileExactCount(input)
	batch, err := newRecordBatch(record, count, api.CountModeExact, input.binding)
	consumer := &batchConsumerFixture{}
	consumeErr := consumeValidatedRecordBatch(batch, consumer)
	if err != nil || consumeErr != nil || len(consumer.roles) != 2 ||
		consumer.roles[0] != planner.StatementClassRecord ||
		consumer.roles[1] != planner.StatementClassExactCount ||
		consumer.record.Validate() != nil || consumer.count.Validate() != nil {
		t.Fatalf("exact batch consumption=%v roles=%v", consumeErr, consumer.roles)
	}
	mode, modeErr := batch.CountMode()
	if modeErr != nil || mode != api.CountModeExact {
		t.Fatal("exact batch mode was not sealed")
	}
	failing := &batchConsumerFixture{failCount: true}
	if !errors.Is(consumeValidatedRecordBatch(batch, failing), errBatchConsumerFixture) ||
		len(failing.roles) != 2 {
		t.Fatal("exact-count consumer failure was swallowed")
	}
	var nilConsumer *batchConsumerFixture
	if !errors.Is(consumeValidatedRecordBatch(batch, nilConsumer), ErrContract) {
		t.Fatal("typed-nil batch consumer was accepted")
	}
}

func TestRecordBatchCountNoneConsumesOnlyRecord(t *testing.T) {
	_, _, query := fixtureSimpleQuery(t, "display", "display")
	input := fixtureRecordInput(t, "forms", query, fixtureAllVisibility(t))
	record, _ := compileRecord(input)
	batch, err := newRecordBatch(
		record, ExactCountStatement{}, api.CountModeNone, input.binding,
	)
	consumer := &batchConsumerFixture{}
	consumeErr := consumeValidatedRecordBatch(batch, consumer)
	if err != nil || consumeErr != nil || len(consumer.roles) != 1 ||
		consumer.roles[0] != planner.StatementClassRecord {
		t.Fatalf("count-none batch roles=%v err=%v", consumer.roles, consumeErr)
	}
}

func TestRecordBatchTamperAndResealCannotDropExactCount(t *testing.T) {
	input, _ := fixtureExactCountScenario(t)
	record, _ := compileRecord(input)
	count, _ := compileExactCount(input)
	batch, _ := newRecordBatch(record, count, api.CountModeExact, input.binding)

	tampered := batch
	tampered.exactCount = ExactCountStatement{}
	tampered.hasExactCount = false
	tampered.stateSeal, _ = recordBatchSeal(tampered)
	if !errors.Is(tampered.Validate(), ErrContract) {
		t.Fatal("resealed batch silently dropped exact count")
	}
	tampered = batch
	tampered.countMode = api.CountModeNone
	tampered.exactCount = ExactCountStatement{}
	tampered.hasExactCount = false
	tampered.stateSeal, _ = recordBatchSeal(tampered)
	if !errors.Is(tampered.Validate(), ErrContract) {
		t.Fatal("resealed exact batch downgraded to count none")
	}
	tampered = batch
	tampered.exactCount.class = planner.StatementClassRecord
	tampered.exactCount.stateSeal, _ = exactCountStatementSeal(tampered.exactCount)
	tampered.stateSeal, _ = recordBatchSeal(tampered)
	if !errors.Is(tampered.Validate(), ErrContract) {
		t.Fatal("resealed wrong count class was accepted")
	}
	tampered = batch
	tampered.binding.continuation[0] ^= 0xff
	tampered.stateSeal, _ = recordBatchSeal(tampered)
	if !errors.Is(tampered.Validate(), ErrContract) {
		t.Fatal("resealed transplanted batch binding was accepted")
	}
}

func TestRecordBatchRedactsAndRejectsJSON(t *testing.T) {
	input, _ := fixtureExactCountScenario(t)
	record, _ := compileRecord(input)
	count, _ := compileExactCount(input)
	batch, _ := newRecordBatch(record, count, api.CountModeExact, input.binding)
	if fmt.Sprint(batch) != compilerRedaction {
		t.Fatal("record batch formatting leaked")
	}
	if document, err := json.Marshal(batch); err == nil || document != nil {
		t.Fatal("record batch JSON marshal did not fail closed")
	}
}
