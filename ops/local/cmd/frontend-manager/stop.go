package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

func runStop(cfg runtimeConfig) error {
	receipt, err := readReceipt(cfg.ReceiptPath)
	if errors.Is(err, errReceiptNotFound) {
		if frontendReady(context.Background(), cfg.ReadyURL) {
			fmt.Println("frontend-v3: unmanaged server is running on port 5173; left untouched")
		} else {
			fmt.Println("frontend-v3: already stopped")
		}
		return nil
	}
	if err != nil {
		return err
	}
	if err := validateReceipt(cfg, receipt); err != nil {
		return err
	}
	alive, ownershipErr := ownedProcessAlive(receipt)
	if ownershipErr != nil {
		fmt.Printf("frontend-v3: WARNING: %v; stale receipt retained\n", ownershipErr)
		return nil
	}
	if alive {
		if err := terminateProcessTree(receipt.PID); err != nil {
			return err
		}
		if err := waitUntilStopped(receipt); err != nil {
			return err
		}
	}
	if err := os.Remove(cfg.ReceiptPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove frontend receipt: %w", err)
	}
	fmt.Println("frontend-v3: stopped")
	return nil
}

func waitUntilStopped(receipt processReceipt) error {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		alive, err := ownedProcessAlive(receipt)
		if err != nil || !alive {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("frontend process PID %d did not stop", receipt.PID)
}
