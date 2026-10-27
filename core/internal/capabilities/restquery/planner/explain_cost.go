package planner

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
)

func decodeCost(document []byte) (Cost, error) {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return Cost{}, ErrExplainContract
	}
	number, numeric := decoded.(json.Number)
	if !numeric || strings.HasPrefix(number.String(), "-") {
		return Cost{}, ErrExplainContract
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Cost{}, ErrExplainContract
	}
	value, err := strconv.ParseFloat(number.String(), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return Cost{}, ErrExplainContract
	}
	return Cost{value: value}, nil
}
