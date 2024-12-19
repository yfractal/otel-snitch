package snitchreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

type snitchReceiver struct {
	host         component.Host
	cancel       context.CancelFunc
	logger       *zap.Logger
	nextConsumer consumer.Traces
	config       *Config
}

func (receiver *snitchReceiver) Start(ctx context.Context, host component.Host) error {
	receiver.host = host
	ctx = context.Background()
	ctx, receiver.cancel = context.WithCancel(ctx)

	return nil
}

func (tailtracerRcvr *snitchReceiver) Shutdown(ctx context.Context) error {
	if tailtracerRcvr.cancel != nil {
		tailtracerRcvr.cancel()
	}
	return nil
}
