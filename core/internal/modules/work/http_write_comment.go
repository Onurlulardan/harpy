package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runtime *PlatformHTTPRuntime) writeComment(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire,
) (ports.OperationReceipt, bool, error) {
	switch input.Operation {
	case "createWorkComment", "createLegacyTaskComment":
		taskID, err := routeID(input.IDs, "task_id")
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		stores, err := runtime.stores(ctx, scope)
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		task, err := stores.Tasks().LoadTaskForUpdate(ctx, taskID)
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		receipt, err := runtime.services.comments.CreateComment(ctx, scope,
			mutation(input, 0), task.BoardID, taskID,
			taskCommentBody(wire, input.Compatibility))
		return receipt, true, err
	case "redactWorkComment", "deleteLegacyTaskComment":
		commentID, err := routeID(input.IDs, "comment_id")
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		version, reason, err := runtime.commentRedactionValues(
			ctx, scope, commentID, wire.ExpectedVersion, wire.Reason, input.Compatibility)
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		receipt, err := runtime.services.comments.RedactComment(ctx, scope,
			mutation(input, version), commentID, reason)
		return receipt, true, err
	}
	return ports.OperationReceipt{}, false, nil
}

func taskCommentBody(wire writeWire, compatibility bool) string {
	if compatibility {
		return wire.Content
	}
	return wire.Body
}

func (runtime *PlatformHTTPRuntime) commentRedactionValues(ctx context.Context,
	scope ports.Scope, commentID ports.ID, version int64, reason string,
	compatibility bool,
) (int64, string, error) {
	if !compatibility || version > 0 {
		return version, reason, nil
	}
	stores, err := runtime.stores(ctx, scope)
	if err != nil {
		return 0, "", err
	}
	comment, err := stores.Collaboration().LoadCommentForUpdate(ctx, commentID)
	if err == nil && comment.AuthorID != scope.ActorID() && reason == "" {
		reason = "legacy compatibility removal"
	}
	return comment.RowVersion, reason, err
}
