package asyncadoptionv1

import (
	"encoding/json"
	"fmt"
	"time"
)

// V1QueueMessage preserves the exact encoding/json wire used by V1 queue.Message.
type V1QueueMessage struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Payload     map[string]any `json:"payload"`
	Priority    uint8          `json:"priority"`
	RetryCount  int            `json:"retry_count"`
	PublishedAt time.Time      `json:"published_at"`
}

type V1IdempotencyResponse struct {
	State       string `json:"state"`
	RequestHash string `json:"request_hash"`
	Status      int    `json:"status"`
}

func DecodeV1QueueMessage(document []byte) (V1QueueMessage, error) {
	var value V1QueueMessage
	if err := decodeV1Wire(document, &value); err != nil {
		return V1QueueMessage{}, err
	}
	if value.ID == "" || value.Type == "" || value.Payload == nil ||
		value.RetryCount < 0 || value.PublishedAt.IsZero() {
		return V1QueueMessage{}, fmt.Errorf("V1 queue message is incomplete")
	}
	return value, nil
}

func DecodeV1IdempotencyResponse(document []byte) (V1IdempotencyResponse, error) {
	var value V1IdempotencyResponse
	if err := decodeV1Wire(document, &value); err != nil {
		return V1IdempotencyResponse{}, err
	}
	validState := value.State == "pending" && value.Status == 0 ||
		value.State == "complete" && value.Status >= 100 && value.Status <= 599
	if !validState || !lowerSHA256.MatchString(value.RequestHash) {
		return V1IdempotencyResponse{}, fmt.Errorf("V1 idempotency response is incomplete")
	}
	return value, nil
}

func decodeV1Wire(document []byte, target any) error {
	if len(document) == 0 || len(document) > MaximumDocumentBytes {
		return fmt.Errorf("V1 wire document size is invalid")
	}
	if err := json.Unmarshal(document, target); err != nil {
		return fmt.Errorf("decode characterized V1 wire: %w", err)
	}
	return nil
}
