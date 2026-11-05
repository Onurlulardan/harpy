package provisionlaunch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

var RequestPath = "/run/installation/provision-request.json"

var ErrPending = errors.New("tenant provisioning is waiting for host execution")

type provisionRequest struct {
	TenantID string `json:"tenant_id"`
}

func Run(ctx context.Context, tenantID string) error {
	if ctx == nil || ctx.Err() != nil {
		return fmt.Errorf("tenant provisioning command is invalid")
	}
	parsed, err := uuid.Parse(tenantID)
	if err != nil || parsed == uuid.Nil {
		return fmt.Errorf("tenant provisioning command is invalid")
	}
	if err = writeProvisionRequest(parsed.String()); err != nil {
		return err
	}
	return ErrPending
}

func ReadTenantID() (string, error) {
	file, err := os.Open(RequestPath)
	if err != nil {
		return "", fmt.Errorf("read tenant provisioning request: %w", err)
	}
	defer file.Close()
	var request provisionRequest
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&request); err != nil {
		return "", fmt.Errorf("read tenant provisioning request: %w", err)
	}
	parsed, err := uuid.Parse(request.TenantID)
	if err != nil || parsed == uuid.Nil || parsed.String() != request.TenantID {
		return "", fmt.Errorf("tenant provisioning request is invalid")
	}
	return parsed.String(), nil
}

func writeProvisionRequest(tenantID string) error {
	file, err := os.CreateTemp(filepath.Dir(RequestPath), ".provision-request-*")
	if err != nil {
		return fmt.Errorf("write tenant provisioning request: %w", err)
	}
	temporary := file.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(temporary)
		}
	}()
	if err = file.Chmod(0o600); err != nil {
		_ = file.Close()
		return fmt.Errorf("write tenant provisioning request: %w", err)
	}
	if err = json.NewEncoder(file).Encode(provisionRequest{TenantID: tenantID}); err != nil {
		_ = file.Close()
		return fmt.Errorf("write tenant provisioning request: %w", err)
	}
	if err = file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("write tenant provisioning request: %w", err)
	}
	if err = file.Close(); err != nil {
		return fmt.Errorf("write tenant provisioning request: %w", err)
	}
	if err = os.Rename(temporary, RequestPath); err != nil {
		return fmt.Errorf("write tenant provisioning request: %w", err)
	}
	committed = true
	return nil
}
