package aiinferencev1

import "math"

func ValidateEmbedding(request EmbeddingRequest) error {
	if err := validateBinding(request.Binding, OperationEmbed); err != nil ||
		len(request.Inputs) < 1 || len(request.Inputs) > MaximumEmbeddingInputs ||
		request.TimeoutMS < MinimumTimeoutMS || request.TimeoutMS > MaximumTimeoutMS ||
		request.Dimensions < 0 || request.Dimensions > MaximumEmbeddingDimensions {
		return ErrInvalidRequest
	}
	total := 0
	for _, input := range request.Inputs {
		total += len(input)
		if !validText(input, MaximumMessageBytes, false) || total > MaximumInputTextBytes {
			return ErrInvalidRequest
		}
	}
	expected, err := CanonicalEmbeddingHash(request)
	if err != nil || !HashMatches(request.RequestHash, expected) {
		return ErrInvalidRequest
	}
	return nil
}

func ValidateVectors(vectors [][]float64, expected, dimensions int) bool {
	if len(vectors) != expected || len(vectors) == 0 {
		return false
	}
	width, total := len(vectors[0]), 0
	if width < 1 || width > MaximumEmbeddingDimensions ||
		(dimensions > 0 && width != dimensions) {
		return false
	}
	for _, vector := range vectors {
		if len(vector) != width {
			return false
		}
		total += len(vector)
		if total > MaximumEmbeddingValues {
			return false
		}
		for _, value := range vector {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return false
			}
		}
	}
	return true
}
