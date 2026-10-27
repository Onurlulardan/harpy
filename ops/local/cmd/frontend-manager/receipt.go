package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var errReceiptNotFound = errors.New("frontend receipt not found")

type processReceipt struct {
	Version         int       `json:"version"`
	PID             int       `json:"pid"`
	ProcessIdentity string    `json:"process_identity"`
	FrontendRoot    string    `json:"frontend_root"`
	ReadyURL        string    `json:"ready_url"`
	StartedAt       time.Time `json:"started_at"`
}

func readReceipt(path string) (processReceipt, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return processReceipt{}, errReceiptNotFound
	}
	if err != nil {
		return processReceipt{}, fmt.Errorf("open frontend receipt: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var receipt processReceipt
	if err := decoder.Decode(&receipt); err != nil {
		return processReceipt{}, fmt.Errorf("decode frontend receipt: %w", err)
	}
	if receipt.Version != 1 || receipt.PID <= 0 || receipt.ProcessIdentity == "" {
		return processReceipt{}, errors.New("frontend receipt has an invalid identity")
	}
	return receipt, nil
}

func writeReceipt(path string, receipt processReceipt) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create frontend state directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), "receipt-*.json")
	if err != nil {
		return fmt.Errorf("create frontend receipt: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("protect frontend receipt: %w", err)
	}
	encoder := json.NewEncoder(temporary)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(receipt); err != nil {
		temporary.Close()
		return fmt.Errorf("encode frontend receipt: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close frontend receipt: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("publish frontend receipt: %w", err)
	}
	return nil
}

func validateReceipt(cfg runtimeConfig, receipt processReceipt) error {
	readyURLMatches := receipt.ReadyURL == cfg.ReadyURL ||
		(cfg.ReadyURL == localFrontendReadyURL &&
			receipt.ReadyURL == legacyLocalFrontendReadyURL)
	if receipt.FrontendRoot != cfg.FrontendRoot || !readyURLMatches {
		return errors.New("frontend receipt does not belong to this workspace")
	}
	return nil
}
