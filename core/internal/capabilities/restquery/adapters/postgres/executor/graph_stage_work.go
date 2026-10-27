package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
)

const maximumGraphWorkStatements = 3

type graphStageWork struct {
	driver     execution.SnapshotDriver
	statements []statement
	run        func(context.Context, execution.SnapshotDriver) error
	stage      int
	ordinal    int
}

func (work graphStageWork) validate() error {
	if interfaceIsNil(work.driver) || work.run == nil ||
		work.stage < 0 || work.stage > 4 || work.ordinal < 0 || work.ordinal > 3 ||
		len(work.statements) < 1 || len(work.statements) > maximumGraphWorkStatements {
		return ErrContract
	}
	for _, candidate := range work.statements {
		if candidate.sql == "" || candidate.explainSQL == "" ||
			!candidate.class.Valid() {
			return ErrContract
		}
	}
	return nil
}

type graphStageSubmission struct {
	work    graphStageWork
	reply   chan error
	waiting bool
}

func (submission graphStageSubmission) validate() error {
	if submission.reply == nil {
		return ErrContract
	}
	if !submission.waiting {
		return submission.work.validate()
	}
	if submission.work.stage < 0 || submission.work.stage > 4 ||
		submission.work.ordinal < 0 || submission.work.ordinal > 3 ||
		!interfaceIsNil(submission.work.driver) || submission.work.run != nil ||
		len(submission.work.statements) != 0 {
		return ErrContract
	}
	return nil
}
