package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func extractRequirements(
	spec resolved.QuerySpec,
) (requirementSet, error) {
	if spec.Validate() != nil {
		return requirementSet{}, errPolicyEvaluationContract
	}
	catalogs, err := spec.Catalogs()
	if err != nil {
		return requirementSet{}, errPolicyEvaluationContract
	}
	root, err := catalogs.RootCatalog()
	if err != nil {
		return requirementSet{}, errPolicyEvaluationContract
	}
	rootPath, err := spec.RootPath()
	if err != nil {
		return requirementSet{}, errPolicyEvaluationContract
	}
	collector := newRequirementCollector()
	if err := collector.addResource(root, rootPath); err != nil {
		return requirementSet{}, err
	}
	mode, err := spec.Mode()
	if err != nil {
		return requirementSet{}, errPolicyEvaluationContract
	}
	switch mode {
	case api.QueryModeRecord:
		record, found, recordErr := spec.Record()
		if recordErr != nil || !found ||
			extractRecordRequirements(
				collector, catalogs, root, rootPath, record,
			) != nil {
			return requirementSet{}, errPolicyEvaluationContract
		}
	case api.QueryModeAggregate:
		aggregate, found, aggregateErr := spec.Aggregate()
		if aggregateErr != nil || !found ||
			extractAggregateRequirements(
				collector, root, rootPath, aggregate,
			) != nil {
			return requirementSet{}, errPolicyEvaluationContract
		}
	default:
		return requirementSet{}, errPolicyEvaluationContract
	}
	return collector.build()
}
