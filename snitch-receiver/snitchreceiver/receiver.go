package snitchreceiver

import (
	"context"
	"net/http"

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
	receiver.startHTTPServer(ctx)

	return nil
}

func (receiver *snitchReceiver) Shutdown(ctx context.Context) error {
	if receiver.cancel != nil {
		receiver.cancel()
	}
	return nil
}

func (receiver *snitchReceiver) startHTTPServer(ctx context.Context) {
	http.HandleFunc("/traces", func(w http.ResponseWriter, r *http.Request) {
		file := r.URL.Query().Get("file")
		if file == "" {
			http.Error(w, "Missing 'file' parameter", http.StatusBadRequest)
			return
		}

		receiver.logger.Info("Received a request for traces.", zap.String("file", file))

		ReadFile(file)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Traces received"))
	})

	server := &http.Server{Addr: ":8081"}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			receiver.logger.Fatal("HTTP server ListenAndServe", zap.Error(err))
		}
	}()

	<-ctx.Done()
	if err := server.Shutdown(context.Background()); err != nil {
		receiver.logger.Fatal("HTTP server Shutdown", zap.Error(err))
	}
}
