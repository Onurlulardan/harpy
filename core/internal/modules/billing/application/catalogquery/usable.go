package catalogquery

import "context"

func (queries *Queries) usable(ctx context.Context) bool {
	return queries != nil && queries.store != nil && ctx != nil && ctx.Err() == nil
}
