package deadletter

import (
	"fmt"
	"regexp"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

const EventRetryPayloadRevision = 1

var eventRetryDomain = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

func EventRetryKind(domain string) (string, error) {
	if !eventRetryDomain.MatchString(domain) {
		return "", fmt.Errorf("dead-letter event retry domain is invalid")
	}
	return domain + ".deadletter_event_retry", nil
}

func EventRetryPolicy(domain string) (jobsv1.KindPolicy, error) {
	kind, err := EventRetryKind(domain)
	if err != nil {
		return jobsv1.KindPolicy{}, err
	}
	return jobsv1.NewKindPolicy(
		kind, EventRetryPayloadRevision, domain,
		jobsv1.EffectExternal, jobsv1.DefaultMaximumAttempts,
	)
}
