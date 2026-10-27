package domain

import (
	"errors"
	"math"
)

const (
	SparseGap       int64 = 1024
	MaxBoardColumns       = 100
	MaxColumnTasks        = 10_000
	MaxBoardTasks         = 50_000
)

var (
	ErrOrderingInvalid = errors.New("work ordering request is invalid")
	ErrBoardLimit      = errors.New("work board limit is exceeded")
)

type RankedTask struct {
	ID       ID
	Position int64
}

type MoveNeighbors struct {
	Before *ID
	After  *ID
}

type PositionPlan struct {
	Position   int64
	Rebalanced []RankedTask
}

func PlanPosition(tasks []RankedTask, moving ID, neighbors MoveNeighbors) (PositionPlan, error) {
	ordered, err := validatedTarget(tasks, moving)
	if err != nil {
		return PositionPlan{}, err
	}
	if len(ordered) >= MaxColumnTasks {
		return PositionPlan{}, ErrBoardLimit
	}
	left, right, err := neighborBounds(ordered, neighbors)
	if err != nil {
		return PositionPlan{}, err
	}
	if position, ok := sparsePosition(left, right); ok {
		return PositionPlan{Position: position}, nil
	}
	rebalanced := make([]RankedTask, len(ordered))
	for index, task := range ordered {
		task.Position = int64(index+1) * SparseGap
		rebalanced[index] = task
	}
	left, right, err = neighborBounds(rebalanced, neighbors)
	if err != nil {
		return PositionPlan{}, err
	}
	position, ok := sparsePosition(left, right)
	if !ok {
		return PositionPlan{}, ErrOrderingInvalid
	}
	return PositionPlan{Position: position, Rebalanced: rebalanced}, nil
}

func validatedTarget(tasks []RankedTask, moving ID) ([]RankedTask, error) {
	if !moving.Valid() || len(tasks) > MaxColumnTasks {
		return nil, ErrOrderingInvalid
	}
	result := make([]RankedTask, 0, len(tasks))
	seen := make(map[ID]struct{}, len(tasks))
	last := int64(0)
	for _, task := range tasks {
		if !task.ID.Valid() || task.Position <= last {
			return nil, ErrOrderingInvalid
		}
		last = task.Position
		if _, exists := seen[task.ID]; exists {
			return nil, ErrOrderingInvalid
		}
		seen[task.ID] = struct{}{}
		if task.ID != moving {
			result = append(result, task)
		}
	}
	return result, nil
}

func neighborBounds(tasks []RankedTask, neighbors MoveNeighbors) (int64, int64, error) {
	beforeIndex := indexOf(tasks, neighbors.Before)
	afterIndex := indexOf(tasks, neighbors.After)
	if neighbors.Before != nil && beforeIndex < 0 || neighbors.After != nil && afterIndex < 0 {
		return 0, 0, ErrOrderingInvalid
	}
	if neighbors.Before != nil && neighbors.After != nil && beforeIndex+1 != afterIndex {
		return 0, 0, ErrOrderingInvalid
	}
	if neighbors.Before != nil && neighbors.After != nil {
		return tasks[beforeIndex].Position, tasks[afterIndex].Position, nil
	}
	if neighbors.Before != nil {
		return tasks[beforeIndex].Position, math.MaxInt64, nil
	}
	if neighbors.After != nil {
		return 0, tasks[afterIndex].Position, nil
	}
	if len(tasks) == 0 {
		return 0, SparseGap * 2, nil
	}
	return tasks[len(tasks)-1].Position, math.MaxInt64, nil
}

func indexOf(tasks []RankedTask, wanted *ID) int {
	if wanted == nil {
		return -1
	}
	for index, task := range tasks {
		if task.ID == *wanted {
			return index
		}
	}
	return -1
}

func sparsePosition(left, right int64) (int64, bool) {
	if left < 0 || right <= left {
		return 0, false
	}
	if right == math.MaxInt64 {
		if left <= math.MaxInt64-SparseGap {
			return left + SparseGap, true
		}
		return 0, false
	}
	if right-left <= 1 {
		return 0, false
	}
	return left + (right-left)/2, true
}
