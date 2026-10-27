package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type developmentBindingInputReaderStub struct {
	result ports.DevelopmentBindingInput
}

func (stub developmentBindingInputReaderStub) ReadDevelopmentBindingInput(
	context.Context, string,
) (ports.DevelopmentBindingInput, error) {
	return stub.result, nil
}

func TestReadDevelopmentBindingInputRejectsMissingReader(t *testing.T) {
	_, err := ReadDevelopmentBindingInput(context.Background(), nil, "environment")
	if err == nil {
		t.Fatal("missing development binding input reader was accepted")
	}
}

func TestReadDevelopmentBindingInputReturnsReaderResult(t *testing.T) {
	want := ports.DevelopmentBindingInput{OperationID: "operation"}
	got, err := ReadDevelopmentBindingInput(
		context.Background(), developmentBindingInputReaderStub{result: want}, "environment",
	)
	if err != nil || got != want {
		t.Fatalf("read development binding input = %#v, %v", got, err)
	}
	if errors.Is(err, context.Canceled) {
		t.Fatal("unexpected context error")
	}
}
