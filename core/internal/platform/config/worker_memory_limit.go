package config

import (
	"math"
	"strconv"
)

const (
	workerMemoryLimitKey          = "GOMEMLIMIT"
	workerMinimumMemoryLimitBytes = 512 * 1024 * 1024
)

type WorkerMemoryLimit struct{ bytes int64 }

func loadWorkerMemoryLimit(values environmentValues) (WorkerMemoryLimit, error) {
	raw, present := values.lookup(workerMemoryLimitKey)
	if !present {
		return WorkerMemoryLimit{}, configError(workerMemoryLimitKey, "is required")
	}
	bytes, valid := parseWorkerMemoryBytes(raw)
	if !valid || bytes < workerMinimumMemoryLimitBytes {
		return WorkerMemoryLimit{}, configError(
			workerMemoryLimitKey, "must be at least 512MiB",
		)
	}
	return WorkerMemoryLimit{bytes: bytes}, nil
}

func parseWorkerMemoryBytes(raw string) (int64, bool) {
	if raw == "" {
		return 0, false
	}
	last := raw[len(raw)-1]
	if last >= '0' && last <= '9' {
		return parseWorkerMemoryInteger(raw, 1)
	}
	if last != 'B' || len(raw) < 2 {
		return 0, false
	}
	previous := raw[len(raw)-2]
	if previous >= '0' && previous <= '9' {
		return parseWorkerMemoryInteger(raw[:len(raw)-1], 1)
	}
	if previous != 'i' || len(raw) < 4 {
		return 0, false
	}
	power := 0
	switch raw[len(raw)-3] {
	case 'K':
		power = 1
	case 'M':
		power = 2
	case 'G':
		power = 3
	case 'T':
		power = 4
	default:
		return 0, false
	}
	multiplier := int64(1)
	for range power {
		multiplier *= 1024
	}
	return parseWorkerMemoryInteger(raw[:len(raw)-3], multiplier)
}

func parseWorkerMemoryInteger(raw string, multiplier int64) (int64, bool) {
	if !decimalDigits(raw) {
		return 0, false
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value > math.MaxInt64/multiplier {
		return 0, false
	}
	return value * multiplier, true
}

func (limit WorkerMemoryLimit) Bytes() int64 { return limit.bytes }

func (limit WorkerMemoryLimit) ValidateRuntime(actualBytes int64) error {
	if limit.bytes < workerMinimumMemoryLimitBytes ||
		actualBytes < workerMinimumMemoryLimitBytes {
		return configError(workerMemoryLimitKey, "runtime memory limit is below 512MiB")
	}
	return nil
}
