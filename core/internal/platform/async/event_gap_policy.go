package async

import (
	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func EventGapJobPolicy(domain string) (jobsv1.KindPolicy, error) {
	return asyncmanifestv1.GapJobPolicy(domain)
}
