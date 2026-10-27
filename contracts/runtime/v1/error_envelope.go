package runtimev1

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

const maximumErrorIssues = 20

type ErrorEnvelope struct {
	Success bool         `json:"success"`
	Error   RuntimeError `json:"error"`
}

type RuntimeError struct {
	Code          string  `json:"code"`
	Message       string  `json:"message"`
	Issues        []Issue `json:"issues"`
	CorrelationID string  `json:"correlation_id"`
}

type Issue struct {
	Path string `json:"path"`
	Code string `json:"code"`
}

func NewErrorEnvelope(
	code string,
	message string,
	correlationID string,
	issues []Issue,
) (ErrorEnvelope, error) {
	if issues == nil {
		issues = []Issue{}
	} else {
		issues = append([]Issue(nil), issues...)
	}
	envelope := ErrorEnvelope{
		Success: false,
		Error: RuntimeError{
			Code:          code,
			Message:       message,
			Issues:        issues,
			CorrelationID: correlationID,
		},
	}
	if err := envelope.Validate(); err != nil {
		return ErrorEnvelope{}, err
	}
	return envelope, nil
}

func NewIssue(path string, code string) (Issue, error) {
	issue := Issue{Path: path, Code: code}
	if err := issue.Validate(); err != nil {
		return Issue{}, err
	}
	return issue, nil
}

func (envelope ErrorEnvelope) Validate() error {
	if envelope.Success {
		return fmt.Errorf("runtime error envelope success must be false")
	}
	return envelope.Error.Validate()
}

func (runtimeError RuntimeError) Validate() error {
	if !validMachineCode(runtimeError.Code) {
		return fmt.Errorf("runtime error code is invalid")
	}
	if !validTextLength(runtimeError.Message, 1, 128) {
		return fmt.Errorf("runtime error message is invalid")
	}
	if runtimeError.Issues == nil || len(runtimeError.Issues) > maximumErrorIssues {
		return fmt.Errorf("runtime error issues must contain zero to 20 items")
	}
	for index, issue := range runtimeError.Issues {
		if err := issue.Validate(); err != nil {
			return fmt.Errorf("runtime error issue %d: %w", index, err)
		}
	}
	if !validCorrelationID(runtimeError.CorrelationID) {
		return fmt.Errorf("runtime error correlation ID is invalid")
	}
	return nil
}

func (issue Issue) Validate() error {
	if !validJSONPointer(issue.Path) {
		return fmt.Errorf("issue path is invalid")
	}
	if !validMachineCode(issue.Code) {
		return fmt.Errorf("issue code is invalid")
	}
	return nil
}

func validJSONPointer(value string) bool {
	if !validTextLength(value, 0, 256) {
		return false
	}
	if value == "" {
		return true
	}
	if value[0] != '/' {
		return false
	}
	for index := 1; index < len(value); index++ {
		if value[index] != '~' {
			continue
		}
		index++
		if index == len(value) || value[index] != '0' && value[index] != '1' {
			return false
		}
	}
	return true
}

func (envelope ErrorEnvelope) MarshalJSON() ([]byte, error) {
	if err := envelope.Validate(); err != nil {
		return nil, err
	}
	type wireEnvelope ErrorEnvelope
	return json.Marshal(wireEnvelope(envelope))
}

func validMachineCode(value string) bool {
	if len(value) < 2 || len(value) > 64 || value[0] < 'A' || value[0] > 'Z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if (character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

func validCorrelationID(value string) bool {
	if len(value) < 1 || len(value) > 64 || !asciiAlphaNumeric(value[0]) {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if !asciiAlphaNumeric(character) && character != '.' &&
			character != '_' && character != '-' {
			return false
		}
	}
	return true
}

func asciiAlphaNumeric(character byte) bool {
	return character >= 'A' && character <= 'Z' ||
		character >= 'a' && character <= 'z' ||
		character >= '0' && character <= '9'
}

func validTextLength(value string, minimum int, maximum int) bool {
	return utf8.ValidString(value) && utf8.RuneCountInString(value) >= minimum &&
		utf8.RuneCountInString(value) <= maximum
}
