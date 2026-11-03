package environmentoperator

import (
	"errors"
	"regexp"
	"strings"
)

// FailureDiagnostic is the safe operator failure record. It never carries a
// DSN, password, token, signing key, or raw credential.
type FailureDiagnostic struct {
	Stage     string
	ErrorCode string
	Detail    string
	SQLState  string
}

type stageFailure struct {
	sentinel error
	stage    string
	code     string
	detail   string
	sqlState string
	cause    error
}

// Resumable keeps ErrResumable in the chain and records the failing stage.
func Resumable(stage, code, detail string, cause error) error {
	if code == "" {
		code = "resumable"
	}
	return &stageFailure{
		sentinel: ErrResumable, stage: stage, code: code, detail: detail,
		sqlState: readSQLState(cause), cause: cause,
	}
}

func (failure *stageFailure) Error() string {
	text := failure.sentinel.Error() + ": stage=" + failure.stage
	if failure.code != "" {
		text += " code=" + failure.code
	}
	if failure.sqlState != "" {
		text += " sqlstate=" + failure.sqlState
	}
	if failure.detail != "" {
		text += ": " + redactDiagnostic(failure.detail)
	}
	if failure.cause != nil {
		text += ": " + redactDiagnostic(failure.cause.Error())
	}
	return text
}

func (failure *stageFailure) Unwrap() []error {
	if failure.cause == nil {
		return []error{failure.sentinel}
	}
	return []error{failure.sentinel, failure.cause}
}

func Diagnose(err error) FailureDiagnostic {
	var failure *stageFailure
	if err == nil || !errors.As(err, &failure) {
		return FailureDiagnostic{}
	}
	return FailureDiagnostic{
		Stage: failure.stage, ErrorCode: failure.code,
		Detail: redactDiagnostic(failure.detail), SQLState: failure.sqlState,
	}
}

func ObservedStage(err error) string {
	return Diagnose(err).Stage
}

type sqlStateError interface {
	SQLState() string
}

func readSQLState(err error) string {
	var state sqlStateError
	if err == nil || !errors.As(err, &state) {
		return ""
	}
	code := state.SQLState()
	if len(code) != 5 {
		return ""
	}
	return code
}

var (
	credentialURIPattern = regexp.MustCompile(`(?i)postgres(?:ql)?://\S+`)
	secretAssignPattern  = regexp.MustCompile(`(?i)(password|secret|token|dsn)\s*[:=]\s*\S+`)
	bearerPattern        = regexp.MustCompile(`(?i)bearer\s+\S+`)
)

func redactDiagnostic(value string) string {
	if value == "" {
		return ""
	}
	redacted := credentialURIPattern.ReplaceAllString(value, "postgres://redacted")
	redacted = secretAssignPattern.ReplaceAllString(redacted, "$1=redacted")
	redacted = bearerPattern.ReplaceAllString(redacted, "bearer redacted")
	if strings.Contains(strings.ToLower(redacted), "private key") {
		return "redacted sensitive diagnostic"
	}
	return redacted
}
