package snitchreceiver

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/http"
	"os"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

type Span struct {
	Name                    [256]byte
	TotalRecordedAttributes int32
}

type SnitchReceiver struct {
	host         component.Host
	cancel       context.CancelFunc
	logger       *zap.Logger
	nextConsumer consumer.Traces
	config       *Config
}

func (snitchReceiver *SnitchReceiver) Start(ctx context.Context, host component.Host) error {
	snitchReceiver.host = host
	ctx, snitchReceiver.cancel = context.WithCancel(ctx)
	go snitchReceiver.startHTTPServer(ctx)
	return nil
}

func (snitchReceiver *SnitchReceiver) Shutdown(ctx context.Context) error {
	if snitchReceiver.cancel != nil {
		snitchReceiver.cancel()
	}
	return nil
}

func (snitchReceiver *SnitchReceiver) startHTTPServer(ctx context.Context) {
	http.HandleFunc("/traces", func(w http.ResponseWriter, r *http.Request) {
		file := r.URL.Query().Get("file")
		if file == "" {
			http.Error(w, "Missing 'file' parameter", http.StatusBadRequest)
			return
		}

		snitchReceiver.logger.Info("Received a request for traces.", zap.String("file", file))

		f, err := os.Open(file)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer f.Close()

		for {
			var span Span
			err = binary.Read(f, binary.LittleEndian, &span)
			if err != nil {
				break
			}

			name := string(span.Name[:])
			fmt.Printf("Name: %s, Total Recorded Attributes: %d\n", name, span.TotalRecordedAttributes)
		}

		if err != nil && err.Error() != "EOF" {
			fmt.Println("Error reading file:", err)
		}
		// open the file and mmap the file

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Traces received"))
	})

	server := &http.Server{Addr: ":8081"}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			snitchReceiver.logger.Fatal("HTTP server ListenAndServe", zap.Error(err))
		}
	}()

	<-ctx.Done()
	if err := server.Shutdown(context.Background()); err != nil {
		snitchReceiver.logger.Fatal("HTTP server Shutdown", zap.Error(err))
	}
}
