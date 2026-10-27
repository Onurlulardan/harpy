package api

import "testing"

func TestIntentRequiresCanonicalTypedDataAndClosedChannels(t *testing.T) {
	data := []TemplateValue{{Key: "name", Kind: ValueString, String: "Ada"}}
	digest, ok := CanonicalDataDigest(data)
	if !ok {
		t.Fatal("canonical digest failed")
	}
	request := IntentRequest{Producer: "workflow", NotificationType: "task.due",
		RecipientPrincipalID: testID(1), TemplateVersionID: testID(2),
		LocaleInput: "tr", EffectKey: "instance/node/1", Category: CategoryWorkflowAction,
		Channels: []Channel{ChannelInApp, ChannelEmail}, Data: data, DataDigest: digest,
		CorrelationID: testID(3), CausationID: testID(4)}
	if err := request.Validate(); err != nil {
		t.Fatal(err)
	}
	request.Channels = []Channel{"websocket"}
	if err := request.Validate(); err != ErrIntentInvalid {
		t.Fatalf("unknown channel error = %v", err)
	}
	request.Channels = []Channel{ChannelEmail}
	request.DataDigest[0] ^= 1
	if err := request.Validate(); err != ErrIntentInvalid {
		t.Fatalf("tampered digest error = %v", err)
	}
}

func testID(seed byte) ID {
	value := ID{15: seed}
	return value
}
