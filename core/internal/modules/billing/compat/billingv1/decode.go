package billingv1

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const maximumBodyBytes int64 = 64 * 1024

func decodeBody[T any](r *http.Request) (T, error) {
	var input T
	if r == nil || r.Body == nil {
		return input, ErrRequestInvalid
	}
	reader := &io.LimitedReader{R: r.Body, N: maximumBodyBytes + 1}
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, ErrRequestInvalid
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return input, ErrRequestInvalid
	}
	if reader.N == 0 {
		return input, ErrRequestInvalid
	}
	return input, nil
}

func decodeOptionalBody[T any](r *http.Request) error {
	if r == nil {
		return ErrRequestInvalid
	}
	if r.Body == nil || r.Body == http.NoBody {
		return nil
	}
	_, err := decodeBody[T](r)
	return err
}

func pathID(r *http.Request) (string, error) {
	if r == nil {
		return "", ErrRequestInvalid
	}
	value := r.PathValue("id")
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		return "", ErrRequestInvalid
	}
	return value, nil
}

func required(value string) bool {
	return value != "" && strings.TrimSpace(value) == value
}
