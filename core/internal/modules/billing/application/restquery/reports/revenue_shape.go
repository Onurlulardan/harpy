package reports

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func validateRevenueShape(request api.QueryRequest) error {
	aggregate, found := request.Aggregate()
	if request.Validate() != nil || request.Mode() != api.QueryModeAggregate || !found {
		return revenueShapeError()
	}
	groups := aggregate.GroupBy()
	if len(groups) != 2 || groups[0] != "currency" ||
		groups[1] != "currency_registry_version" {
		return revenueShapeError()
	}
	measures := aggregate.Measures()
	if len(measures) != 2 || !revenueMeasure(
		measures[0], "monthly_minor_units",
	) || !revenueMeasure(measures[1], "annual_minor_units") {
		return revenueShapeError()
	}
	return nil
}

func revenueMeasure(measure api.RequestMeasure, fieldName string) bool {
	field, found := measure.Field()
	return found && field == fieldName && measure.Name() == fieldName &&
		measure.Function() == api.AggregateFunctionSum
}

func revenueShapeError() error {
	issue, err := api.NewQueryIssue("/aggregate", restqueryv1.IssueValueNotAllowed)
	if err != nil {
		return ErrContract
	}
	query, err := api.NewQueryError(restqueryv1.ErrorQueryInvalid, []runtimev1.Issue{issue})
	if err != nil {
		return ErrContract
	}
	return query
}
