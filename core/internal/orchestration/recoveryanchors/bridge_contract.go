package recoveryanchors

import (
	"fmt"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
)

type Service struct {
	anchors  anchorapi.Service
	disabled bool
}

func NewService(anchors anchorapi.Service) (*Service, error) {
	if anchors == nil {
		return nil, fmt.Errorf("recovery anchor service is required")
	}
	return &Service{anchors: anchors}, nil
}

func NewDisabledService() *Service { return &Service{disabled: true} }
