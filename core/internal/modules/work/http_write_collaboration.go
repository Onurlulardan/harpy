package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runtime *PlatformHTTPRuntime) writeCollaboration(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire,
) (ports.OperationReceipt, bool, error) {
	if receipt, matched, err := runtime.writeMembership(
		ctx, scope, input, wire); matched {
		return receipt, true, err
	}
	if receipt, matched, err := runtime.writeCommentLabel(
		ctx, scope, input, wire); matched {
		return receipt, true, err
	}
	if receipt, matched, err := runtime.writeChecklist(
		ctx, scope, input, wire); matched {
		return receipt, true, err
	}
	return runtime.writeCalendar(ctx, scope, input, wire)
}
