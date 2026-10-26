package main

import (
	"context"
	"errors"
	"fmt"
)

func runStatus(cfg runtimeConfig) error {
	receipt, err := readReceipt(cfg.ReceiptPath)
	if errors.Is(err, errReceiptNotFound) {
		if frontendReady(context.Background(), cfg.ReadyURL) {
			return errors.New("unmanaged server is responding on port 5173")
		}
		fmt.Println("frontend-v3: stopped")
		return nil
	}
	if err != nil {
		return err
	}
	if err := validateReceipt(cfg, receipt); err != nil {
		return err
	}
	alive, err := ownedProcessAlive(receipt)
	if err != nil {
		return err
	}
	if !alive {
		return fmt.Errorf("managed PID %d is not running", receipt.PID)
	}
	if !frontendReady(context.Background(), cfg.ReadyURL) {
		return fmt.Errorf("managed PID %d is running but readiness failed", receipt.PID)
	}
	fmt.Printf("frontend-v3: healthy at %s (PID %d)\n", cfg.ReadyURL, receipt.PID)
	return nil
}
