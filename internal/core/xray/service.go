package xray

import (
	"context"

	"bgscan/internal/core/process"
)

// XrayService creates, validates, and starts Xray configurations.
type XrayService interface {
	GetOutboundTemplateByName(string) (*XrayOutboundsFile, error)
	GenerateConfig(outbound, ip string, port uint16) (string, error)
	ValidateConfig(context.Context, string) error
	Start(context.Context, string) (process.Process, error)
}

// xrayService is the default XrayService implementation.
type xrayService struct{}

// NewXrayService creates an Xray service.
func NewXrayService() XrayService {
	return &xrayService{}
}

func (s *xrayService) GetOutboundTemplateByName(name string) (*XrayOutboundsFile, error) {
	return GetOutboundTemplateByName(name)
}

func (s *xrayService) GenerateConfig(outbound, ip string, port uint16) (string, error) {
	return GenerateConfig(outbound, ip, port)
}

func (s *xrayService) ValidateConfig(ctx context.Context, path string) error {
	return ValidateConfig(ctx, path)
}

func (s *xrayService) Start(ctx context.Context, path string) (process.Process, error) {
	return StartXray(ctx, path)
}
