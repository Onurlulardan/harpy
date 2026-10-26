package provider

import (
	"context"
	"errors"
	"reflect"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

var errFixtureRowMissing = errors.New("fixture row missing")

type fixtureRow struct {
	values []any
	err    error
}

func (row fixtureRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != len(row.values) {
		return errors.New("fixture scan width mismatch")
	}
	for index := range destinations {
		target := reflect.ValueOf(destinations[index])
		value := reflect.ValueOf(row.values[index])
		if !target.IsValid() || target.Kind() != reflect.Pointer || target.IsNil() {
			return errors.New("fixture scan target is invalid")
		}
		if row.values[index] == nil {
			target.Elem().Set(reflect.Zero(target.Elem().Type()))
		} else {
			target.Elem().Set(value)
		}
	}
	return nil
}

type fixtureDatabase struct{ rows map[string][]fixtureRow }

func (database *fixtureDatabase) queryRow(
	_ context.Context,
	statement string,
	_ ...any,
) databaseRow {
	rows := database.rows[statement]
	if len(rows) == 0 {
		return fixtureRow{err: errFixtureRowMissing}
	}
	database.rows[statement] = rows[1:]
	return rows[0]
}

type fixtureUpdater struct {
	request subscriptionapp.UpdateFromProviderRequest
	result  subscriptionapp.Result
	calls   int
}

func (updater *fixtureUpdater) UpdateFromProvider(
	_ context.Context,
	request subscriptionapp.UpdateFromProviderRequest,
) (subscriptionapp.Result, error) {
	updater.calls++
	updater.request = request
	return updater.result, nil
}

type fixtureJobs struct {
	ids  []uuid.UUID
	jobs []providerapp.AsyncJob
}

func (jobs *fixtureJobs) EnqueueProviderJob(
	_ context.Context,
	id uuid.UUID,
	job providerapp.AsyncJob,
) error {
	jobs.ids = append(jobs.ids, id)
	jobs.jobs = append(jobs.jobs, job)
	return nil
}
