package postgres

import (
	"context"
	"time"
)

func (session *schemaExecutorTenantSession) close(timeout time.Duration) {
	if session == nil || session.connection == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	var tenantID string
	var released bool
	err := session.connection.QueryRow(
		ctx, schemaExecutorReleaseTenantLockSQL, session.tenantID,
	).Scan(&tenantID, &released)
	cancel()
	if err == nil && tenantID == session.tenantID && released {
		session.connection.Release()
	} else {
		connection := session.connection.Hijack()
		closeCtx, cancelClose := context.WithTimeout(context.Background(), timeout)
		_ = connection.Close(closeCtx)
		cancelClose()
	}
	session.connection = nil
}
