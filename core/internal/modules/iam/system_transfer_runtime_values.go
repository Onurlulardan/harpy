package iam

import (
	"time"

	"github.com/google/uuid"
)

func (runtime *SystemTransferRuntime) now() (time.Time, error) {
	if runtime == nil || runtime.clock == nil {
		return time.Time{}, ErrSystemTransferUnavailable
	}
	value := runtime.clock().UTC().Truncate(time.Microsecond)
	if value.IsZero() {
		return time.Time{}, ErrSystemTransferUnavailable
	}
	return value, nil
}

func (runtime *SystemTransferRuntime) newV7() (uuid.UUID, error) {
	if runtime == nil || runtime.newID == nil {
		return uuid.Nil, ErrSystemTransferUnavailable
	}
	value, err := runtime.newID()
	if err != nil || value == uuid.Nil || value.Version() != 7 ||
		value.Variant() != uuid.RFC4122 {
		return uuid.Nil, ErrSystemTransferUnavailable
	}
	return value, nil
}

func (runtime *SystemTransferRuntime) threeV7() (
	uuid.UUID, uuid.UUID, uuid.UUID, error,
) {
	first, err := runtime.newV7()
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	second, err := runtime.newV7()
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	third, err := runtime.newV7()
	return first, second, third, err
}
