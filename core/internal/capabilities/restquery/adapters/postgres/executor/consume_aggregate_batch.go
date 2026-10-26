package executor

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type consumedAggregate struct {
	statement compiler.AggregateStatement
}

func (batch consumedAggregate) Validate() error {
	if batch.statement.Validate() != nil {
		return ErrContract
	}
	return nil
}

type aggregateConsumer struct {
	result *consumedAggregate
	used   bool
}

func consumeAggregateBatch(
	batch compiler.AggregateBatch,
	plan resolved.AuthorizedPlan,
) (consumedAggregate, error) {
	result := consumedAggregate{}
	consumer := &aggregateConsumer{result: &result}
	if compiler.ConsumeAggregateBatch(batch, plan, consumer) != nil ||
		!consumer.used || result.Validate() != nil {
		return consumedAggregate{}, ErrContract
	}
	return result, nil
}

func (consumer *aggregateConsumer) ConsumeAggregate(
	statement compiler.AggregateStatement,
) error {
	if consumer == nil || consumer.result == nil || consumer.used ||
		statement.Validate() != nil {
		return ErrContract
	}
	consumer.result.statement = statement
	consumer.used = true
	return nil
}
