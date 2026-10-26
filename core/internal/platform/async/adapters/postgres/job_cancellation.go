package postgres

type JobCancellation struct {
	code   string
	digest string
}

func NewJobCancellation(code string, digest string) (JobCancellation, error) {
	if !jobErrorCodePattern.MatchString(code) ||
		code == "JOB_OUTCOME_UNKNOWN" || !claimDigestPattern.MatchString(digest) {
		return JobCancellation{}, ErrJobFailureInvalid
	}
	return JobCancellation{code: code, digest: digest}, nil
}

func (cancellation JobCancellation) valid() bool {
	return jobErrorCodePattern.MatchString(cancellation.code) &&
		cancellation.code != "JOB_OUTCOME_UNKNOWN" &&
		claimDigestPattern.MatchString(cancellation.digest)
}
