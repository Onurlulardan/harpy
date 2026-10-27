package queryexecution

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	MaximumQueries      = 30
	MaximumConcurrency  = 6
	MaximumResponseSize = 5 << 20
	ExecutionTimeout    = 10 * time.Second
	responseDataBudget  = MaximumResponseSize - MaximumQueries*256 - 512
)

var (
	ErrInvalid  = errors.New("analytics query batch is invalid")
	ErrTooLarge = errors.New("analytics query batch response is too large")
	sourceID    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
)

type Item struct {
	DataSourceID       string
	QueryID, VersionID uuid.UUID
	Parameters         []byte
}

type Request struct {
	Items []Item
}

type ItemResult struct {
	DataSourceID string
	Body         []byte
	ErrorCode    string
}

type Result struct {
	Items []ItemResult
}

func Execute(ctx context.Context, request Request, runner Runner) (Result, error) {
	if ctx == nil {
		return Result{}, ErrInvalid
	}
	if err := checkRequest(request, runner); err != nil {
		return Result{}, err
	}
	executionContext, cancel := context.WithTimeout(ctx, ExecutionTimeout)
	defer cancel()

	items := make([]ItemResult, len(request.Items))
	jobs := make(chan int)
	var workers sync.WaitGroup
	for range min(MaximumConcurrency, len(request.Items)) {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range jobs {
				items[index] = run(executionContext, request.Items[index], runner)
			}
		}()
	}
	for index := range request.Items {
		jobs <- index
	}
	close(jobs)
	workers.Wait()

	size := 0
	for index := range items {
		size += len(items[index].DataSourceID) + len(items[index].Body) + len(items[index].ErrorCode)
		if size > responseDataBudget {
			clearResult(items)
			return Result{}, ErrTooLarge
		}
	}
	return Result{Items: items}, nil
}

func run(ctx context.Context, item Item, runner Runner) ItemResult {
	result := ItemResult{DataSourceID: item.DataSourceID}
	if ctx.Err() != nil {
		result.ErrorCode = "QUERY_TIMEOUT"
		return result
	}
	body, err := runner.Run(ctx, item)
	if err != nil {
		if ctx.Err() != nil {
			result.ErrorCode = "QUERY_TIMEOUT"
		} else {
			result.ErrorCode = "QUERY_FAILED"
		}
		return result
	}
	result.Body = append([]byte(nil), body...)
	return result
}

func checkRequest(request Request, runner Runner) error {
	if runner == nil || len(request.Items) == 0 || len(request.Items) > MaximumQueries {
		return ErrInvalid
	}
	seen := make(map[string]struct{}, len(request.Items))
	for _, item := range request.Items {
		if !sourceID.MatchString(item.DataSourceID) || item.QueryID.Version() != 7 ||
			item.VersionID.Version() != 7 || !json.Valid(item.Parameters) {
			return ErrInvalid
		}
		if _, exists := seen[item.DataSourceID]; exists {
			return ErrInvalid
		}
		seen[item.DataSourceID] = struct{}{}
	}
	return nil
}

func clearResult(items []ItemResult) {
	for index := range items {
		clear(items[index].Body)
	}
}
