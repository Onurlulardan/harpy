package executor

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type consumedBatch struct {
	record        compiler.RecordStatement
	exactCount    compiler.ExactCountStatement
	countMode     api.CountMode
	hasExactCount bool
}

type batchConsumer struct {
	batch *consumedBatch
	phase uint8
}

func consumeBatch(
	batch compiler.RecordBatch,
	plan resolved.AuthorizedPlan,
) (consumedBatch, error) {
	mode, err := batch.CountMode()
	if err != nil {
		return consumedBatch{}, ErrContract
	}
	result := consumedBatch{countMode: mode}
	consumer := &batchConsumer{batch: &result}
	if compiler.ConsumeRecordBatch(batch, plan, consumer) != nil ||
		consumer.validCompletion(mode) != nil || result.Validate() != nil {
		return consumedBatch{}, ErrContract
	}
	return result, nil
}

func (consumer *batchConsumer) ConsumeRecord(
	statement compiler.RecordStatement,
) error {
	if consumer == nil || consumer.batch == nil || consumer.phase != 0 ||
		statement.Validate() != nil {
		return ErrContract
	}
	consumer.batch.record = statement
	consumer.phase = 1
	return nil
}

func (consumer *batchConsumer) ConsumeExactCount(
	statement compiler.ExactCountStatement,
) error {
	if consumer == nil || consumer.batch == nil || consumer.phase != 1 ||
		statement.Validate() != nil {
		return ErrContract
	}
	consumer.batch.exactCount = statement
	consumer.batch.hasExactCount = true
	consumer.phase = 2
	return nil
}

func (consumer *batchConsumer) validCompletion(mode api.CountMode) error {
	if consumer == nil || consumer.batch == nil ||
		mode == api.CountModeNone && consumer.phase != 1 ||
		mode == api.CountModeExact && consumer.phase != 2 {
		return ErrContract
	}
	return nil
}
