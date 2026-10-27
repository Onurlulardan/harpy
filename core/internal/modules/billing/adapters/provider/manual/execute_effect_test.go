package manual

import (
	"context"
	"testing"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

func TestExecuteEffectReturnsDeterministicManualReceipt(t *testing.T) {
	adapter, err := New(manualClock{manualNow})
	if err != nil {
		t.Fatal(err)
	}
	request := manualRequest(t, CommandOperation)
	first, err := adapter.ExecuteEffect(context.Background(), request)
	if err != nil || first.IsUnknown() {
		t.Fatalf("result=%+v error=%v", first, err)
	}
	second, err := adapter.ExecuteEffect(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	firstReceipt, firstOK := first.Receipt()
	secondReceipt, secondOK := second.Receipt()
	if !firstOK || !secondOK || firstReceipt.ReceiptID != secondReceipt.ReceiptID ||
		firstReceipt.Code != "MANUAL_COMMAND_ACCEPTED" ||
		firstReceipt.ResultReference == nil ||
		*firstReceipt.ResultReference != request.PayloadReference {
		t.Fatalf("manual receipts differ: %+v %+v", firstReceipt, secondReceipt)
	}
	wrong := manualRequest(t, ReconciliationOperation)
	if _, err := adapter.ExecuteEffect(context.Background(), wrong); err != providerapp.ErrEffectResult {
		t.Fatalf("wrong operation error=%v", err)
	}
}

func TestExecuteEffectRejectsTypedNilClock(t *testing.T) {
	var clock *manualClock
	if _, err := New(clock); err != providerapp.ErrConfiguration {
		t.Fatalf("typed nil clock error=%v", err)
	}
}
