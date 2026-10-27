package application

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func ReadDevelopmentBindingInput(
	ctx context.Context,
	reader ports.DevelopmentBindingInputReader,
	environmentID string,
) (ports.DevelopmentBindingInput, error) {
	if ctx == nil || ctx.Err() != nil || reader == nil || environmentID == "" {
		return ports.DevelopmentBindingInput{}, fmt.Errorf(
			"development authority binding input dependencies are invalid",
		)
	}
	return reader.ReadDevelopmentBindingInput(ctx, environmentID)
}
