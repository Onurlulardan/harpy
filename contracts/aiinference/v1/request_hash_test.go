package aiinferencev1

import (
	"encoding/json"
	"math"
	"testing"
)

func TestSemanticHashIgnoresRandomizedCredentialEnvelope(t *testing.T) {
	request := generationFixture(t)
	request.CredentialEnvelope = json.RawMessage(`{"ciphertext":"first"}`)
	first, err := CanonicalGenerationHash(request)
	if err != nil {
		t.Fatal(err)
	}
	request.CredentialEnvelope = json.RawMessage(`{"ciphertext":"second"}`)
	request.SignedContext = json.RawMessage(`{"signature":"second"}`)
	second, err := CanonicalGenerationHash(request)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("randomized sealed credential changed semantic request hash")
	}
	request.CredentialRevision++
	third, _ := CanonicalGenerationHash(request)
	if first == third {
		t.Fatal("credential revision was not bound")
	}
}

func TestSemanticHashBindsProviderModelConfigInvocationAndReceipt(t *testing.T) {
	original := generationFixture(t)
	want, _ := CanonicalGenerationHash(original)
	cases := []struct {
		name   string
		mutate func(*GenerationRequest)
	}{
		{"provider", func(value *GenerationRequest) { value.Provider = "anthropic" }},
		{"model", func(value *GenerationRequest) { value.Model = "gpt-5.1" }},
		{"config revision", func(value *GenerationRequest) { value.ConfigRevision++ }},
		{"credential revision", func(value *GenerationRequest) { value.CredentialRevision++ }},
		{"invocation", func(value *GenerationRequest) {
			value.InvocationID = "0198eac5-5ee9-7e0f-8c4a-8493a072f899"
		}},
		{"receipt", func(value *GenerationRequest) {
			value.ReceiptID = "0198eac5-5ee9-7e0f-8c4a-8493a072f899"
		}},
	}
	for _, item := range cases {
		changed := original
		item.mutate(&changed)
		got, _ := CanonicalGenerationHash(changed)
		if got == want {
			t.Fatalf("%s was not bound to semantic hash", item.name)
		}
	}
}

func TestGenerationValidationRejectsNonFiniteAndHashReuse(t *testing.T) {
	request := generationFixture(t)
	request.RequestHash, _ = CanonicalGenerationHash(request)
	if err := ValidateGeneration(request, OperationGenerate); err != nil {
		t.Fatal(err)
	}
	request.Messages[0].Content = "changed"
	if err := ValidateGeneration(request, OperationGenerate); err == nil {
		t.Fatal("same hash accepted for changed input")
	}
}

func TestGenerationValidationRejectsNonFiniteTemperatureAndLimits(t *testing.T) {
	request := generationFixture(t)
	notANumber := math.NaN()
	request.Settings.Temperature = &notANumber
	request.RequestHash, _ = CanonicalGenerationHash(request)
	if err := ValidateGeneration(request, OperationGenerate); err == nil {
		t.Fatal("non-finite temperature accepted")
	}
	request = generationFixture(t)
	request.Messages = make([]Message, MaximumMessages+1)
	request.RequestHash, _ = CanonicalGenerationHash(request)
	if err := ValidateGeneration(request, OperationGenerate); err == nil {
		t.Fatal("message limit accepted")
	}
}

func TestAnthropicEmbeddingIsAValidRequestButUnsupportedProviderCapability(t *testing.T) {
	binding := generationFixture(t).Binding
	binding.Operation, binding.Provider, binding.Model = OperationEmbed, "anthropic", "claude-4"
	request := EmbeddingRequest{Binding: binding, Inputs: []string{"hello"}, TimeoutMS: 5000}
	request.RequestHash, _ = CanonicalEmbeddingHash(request)
	if err := ValidateEmbedding(request); err != nil {
		t.Fatalf("capability negotiation was collapsed into request validation: %v", err)
	}
}

func TestEmbeddingVectorsRejectNonFiniteAndShapeDrift(t *testing.T) {
	if !ValidateVectors([][]float64{{1, 2}, {3, 4}}, 2, 2) {
		t.Fatal("finite embedding rejected")
	}
	if ValidateVectors([][]float64{{1, 2}, {3}}, 2, 0) ||
		ValidateVectors([][]float64{{1, 2}, {3, math.Inf(1)}}, 2, 2) {
		t.Fatal("invalid embedding accepted")
	}
}

func generationFixture(t *testing.T) GenerationRequest {
	t.Helper()
	request := GenerationRequest{Binding: Binding{ContractVersion: RequestContractVersion,
		InvocationID:   "0198eac5-5ee9-7e0f-8c4a-8493a072f801",
		ReceiptID:      "0198eac5-5ee9-7e0f-8c4a-8493a072f802",
		TenantID:       "018f936a-0d2d-7d1a-bdf1-9834f1f6a100",
		ConfigID:       "018f936a-0d2d-7d1a-bdf1-9834f1f6a101",
		ConfigRevision: 7, CredentialRevision: 9, AuthorizationVersion: 11,
		Operation: OperationGenerate, Provider: "openai", Model: "gpt-5",
		CredentialEnvelope: json.RawMessage(`{"ciphertext":"sealed"}`),
		SignedContext:      json.RawMessage(`{"token":true}`)},
		Messages: []Message{{Role: "user", Content: "hello"}},
		Settings: GenerationSettings{MaximumOutputTokens: 256, TimeoutMS: 5000}}
	return request
}
