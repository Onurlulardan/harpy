package executor

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestAggregateOwnerAndPlanTransplantsFailClosed(t *testing.T) {
	fixture := newAggregateFixture(t, true)
	batch := consumedAggregateFixture(t, fixture)
	driver, _ := readyAggregateDriver(t, batch, nil, "1")
	driver.owner = "crm"
	if _, err := runAggregate(
		context.Background(), driver, batch, planner.HardPolicy(),
	); !errors.Is(err, ErrContract) || len(driver.events) != 0 {
		t.Fatalf("owner transplant error=%v events=%v", err, driver.events)
	}
	other := newAggregateFixture(t, false)
	if fixture.batch.ValidatePlan(other.plan) == nil {
		t.Fatal("aggregate plan transplant was accepted")
	}
	if _, err := consumeAggregateBatch(
		fixture.batch, other.plan,
	); !errors.Is(err, ErrContract) {
		t.Fatalf("plan-transplanted aggregate batch error=%v", err)
	}
}
