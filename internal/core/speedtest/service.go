package speedtest

import "context"

type Service interface {
	MeasureLatency(context.Context, LatencyConfig) (LatencyResult, error)
	MeasureDownloadSpeed(context.Context, DownloadConfig) (SpeedResult, error)
	MeasureUploadSpeed(context.Context, UploadConfig) (SpeedResult, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) MeasureLatency(
	ctx context.Context,
	cfg LatencyConfig,
) (LatencyResult, error) {
	return measureLatency(ctx, cfg)
}

func (s *service) MeasureDownloadSpeed(
	ctx context.Context,
	cfg DownloadConfig,
) (SpeedResult, error) {
	return measureDownloadSpeed(ctx, cfg)
}

func (s *service) MeasureUploadSpeed(
	ctx context.Context,
	cfg UploadConfig,
) (SpeedResult, error) {
	return measureUploadSpeed(ctx, cfg)
}
