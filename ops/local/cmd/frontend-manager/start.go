package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func runStart(ctx context.Context, cfg runtimeConfig) error {
	receipt, err := readReceipt(cfg.ReceiptPath)
	if err == nil {
		if err := validateReceipt(cfg, receipt); err != nil {
			return err
		}
		alive, err := ownedProcessAlive(receipt)
		if err != nil {
			return err
		}
		if alive && frontendReady(ctx, cfg.ReadyURL) {
			fmt.Printf("frontend-v3: already running at %s (PID %d)\n", cfg.ReadyURL, receipt.PID)
			return nil
		}
		if alive {
			return fmt.Errorf("managed PID %d is running but frontend is not ready", receipt.PID)
		}
		if err := os.Remove(cfg.ReceiptPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove stale frontend receipt: %w", err)
		}
	} else if !errors.Is(err, errReceiptNotFound) {
		return err
	}

	if frontendReady(ctx, cfg.ReadyURL) {
		return errors.New("port 5173 is served by an unmanaged process; it was left untouched")
	}
	return startFrontendProcess(ctx, cfg)
}

func startFrontendProcess(ctx context.Context, cfg runtimeConfig) error {
	yarnPath, err := exec.LookPath("yarn")
	if err != nil {
		return errors.New("yarn was not found; install Yarn and retry task local:up")
	}
	if err := os.MkdirAll(cfg.StateDir, 0o700); err != nil {
		return fmt.Errorf("create frontend state directory: %w", err)
	}
	logFile, err := os.OpenFile(cfg.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open frontend log: %w", err)
	}
	defer logFile.Close()
	fmt.Fprintf(logFile, "\n[%s] starting frontend-v3\n", time.Now().UTC().Format(time.RFC3339))

	command := exec.Command(yarnPath, "dev")
	command.Dir = cfg.FrontendRoot
	command.Stdout = logFile
	command.Stderr = logFile
	command.Env = setEnvironment(os.Environ(), "VITE_API_BASE_URL", "")
	command.Env = setEnvironment(command.Env, "LOCAL_API_PORT", cfg.LocalAPIPort)
	configureManagedProcess(command)
	if err := command.Start(); err != nil {
		return fmt.Errorf("start yarn dev: %w", err)
	}

	identity, alive := processIdentity(command.Process.Pid)
	if !alive {
		return errors.New("frontend process exited before its identity was recorded")
	}
	receipt := processReceipt{
		Version:         1,
		PID:             command.Process.Pid,
		ProcessIdentity: identity,
		FrontendRoot:    cfg.FrontendRoot,
		ReadyURL:        cfg.ReadyURL,
		StartedAt:       time.Now().UTC(),
	}
	if err := writeReceipt(cfg.ReceiptPath, receipt); err != nil {
		_ = terminateProcessTree(command.Process.Pid)
		return err
	}
	if err := waitForFrontend(ctx, cfg, command); err != nil {
		_ = terminateProcessTree(command.Process.Pid)
		_ = os.Remove(cfg.ReceiptPath)
		return fmt.Errorf("%w; inspect %s", err, cfg.LogPath)
	}
	fmt.Printf("frontend-v3: ready at %s (PID %d, API proxy :%s)\n", cfg.ReadyURL, receipt.PID, cfg.LocalAPIPort)
	return nil
}
