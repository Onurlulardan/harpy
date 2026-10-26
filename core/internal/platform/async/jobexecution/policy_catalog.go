package jobexecution

import jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"

type Policy = jobsv1.KindPolicy

func NewPolicyCatalog(policies []Policy) (jobsv1.Catalog, error) {
	return jobsv1.NewCatalog(policies)
}
