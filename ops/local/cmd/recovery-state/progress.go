package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

var progressStates = map[string]bool{
	"claimed": true, "ready_to_stop": true, "restoring": true,
	"verifying": true, "switched": true, "completed": true,
	"interrupted": true, "fenced": true,
}

func runWaitReady(arguments []string) error {
	flags := flag.NewFlagSet("wait-ready", flag.ContinueOnError)
	directory := flags.String("state-directory", "/run/restore-state", "restore state")
	timeout := flags.Duration("timeout", 10*time.Minute, "maximum wait")
	poll := flags.Duration("poll", 500*time.Millisecond, "poll interval")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 ||
		*timeout <= 0 || *poll <= 0 || !filepath.IsAbs(*directory) {
		return fmt.Errorf("wait-ready arguments are invalid")
	}
	deadline := time.Now().Add(*timeout)
	path := filepath.Join(*directory, "progress.json")
	for {
		_, statErr := os.Lstat(path)
		if statErr != nil && !os.IsNotExist(statErr) {
			return fmt.Errorf("inspect restore progress: %w", statErr)
		}
		if statErr == nil {
			progress, err := readProgress(path)
			if err != nil {
				return err
			}
			switch progress.State {
			case "ready_to_stop":
				return nil
			case "claimed":
			case "interrupted", "fenced":
				return fmt.Errorf("restore entered terminal state %s", progress.State)
			default:
				return fmt.Errorf("restore advanced before drain acknowledgement")
			}
		}
		if !time.Now().Before(deadline) {
			return fmt.Errorf("restore did not become ready_to_stop")
		}
		time.Sleep(*poll)
	}
}

func readProgress(path string) (restoreProgress, error) {
	content, err := readRegular(path, 64*1024)
	if err != nil {
		return restoreProgress{}, err
	}
	var progress restoreProgress
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&progress) != nil || requireJSONEOF(decoder) != nil ||
		!validProgress(progress) {
		return restoreProgress{}, fmt.Errorf("restore progress is invalid")
	}
	return progress, nil
}

func validProgress(value restoreProgress) bool {
	operation, operationErr := uuid.Parse(value.OperationID)
	attempt, attemptErr := uuid.Parse(value.AttemptID)
	updatedAt, timeErr := time.Parse(time.RFC3339Nano, value.UpdatedAt)
	return value.Schema == "restore-progress.v1" &&
		operationErr == nil && operation != uuid.Nil && operation.String() == value.OperationID &&
		attemptErr == nil && attempt != uuid.Nil && attempt.String() == value.AttemptID &&
		sha256Pattern.MatchString(value.PlanSHA256) && progressStates[value.State] &&
		timeErr == nil && updatedAt.Location() == time.UTC
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err == io.EOF {
		return nil
	}
	return fmt.Errorf("trailing JSON data")
}
