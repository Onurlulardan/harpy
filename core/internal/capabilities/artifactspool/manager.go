package artifactspool

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	productionReservationBytes int64 = 2_147_483_648
	productionNodeQuotaBytes   int64 = 8_589_934_592
)

type Manager struct {
	root             string
	reservationBytes int64
	slots            int
}

func NewManager(root string) (*Manager, error) {
	return newManager(root, productionReservationBytes, productionNodeQuotaBytes)
}

func newManager(root string, reservationBytes, quotaBytes int64) (*Manager, error) {
	if !filepath.IsAbs(root) || reservationBytes < 1 || quotaBytes < reservationBytes ||
		quotaBytes%reservationBytes != 0 || quotaBytes/reservationBytes > 64 {
		return nil, fmt.Errorf("artifact spool configuration is invalid")
	}
	clean := filepath.Clean(root)
	if err := os.MkdirAll(clean, 0o700); err != nil {
		return nil, fmt.Errorf("create artifact spool directory: %w", err)
	}
	info, err := os.Lstat(clean)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("artifact spool directory is unsafe")
	}
	if err := os.Chmod(clean, 0o700); err != nil {
		return nil, fmt.Errorf("restrict artifact spool directory: %w", err)
	}
	return &Manager{root: clean, reservationBytes: reservationBytes,
		slots: int(quotaBytes / reservationBytes)}, nil
}

func (manager *Manager) lockPath(slot int) string {
	return filepath.Join(manager.root, fmt.Sprintf(".fcart-slot-%02d.lock", slot))
}

func (manager *Manager) spoolPath(slot int) string {
	return filepath.Join(manager.root, fmt.Sprintf("fcart-slot-%02d.spool", slot))
}

func (manager *Manager) identityPath(identity ReservationIdentity) string {
	return filepath.Join(manager.root, fmt.Sprintf(".fcart-attempt-%s-%d.lock",
		identity.ExportID.String(), identity.AttemptGeneration))
}
