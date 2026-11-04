//go:build integration

package foundation

import "context"

const sharedTestClusterLock int64 = 84218421

func LockSharedTestCluster(dsn string) (func(), error) {
	ctx := context.Background()
	conn, err := connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if _, err = conn.Exec(ctx, "SELECT pg_advisory_lock($1)", sharedTestClusterLock); err != nil {
		_ = conn.Close(ctx)
		return nil, err
	}
	return func() {
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", sharedTestClusterLock)
		_ = conn.Close(context.Background())
	}, nil
}
