package tailtracer

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

type Span struct {
	Name                    [256]byte
	TotalRecordedAttributes int32
}

type tailtracerReceiver struct {
	host         component.Host
	cancel       context.CancelFunc
	logger       *zap.Logger
	nextConsumer consumer.Traces
	config       *Config
}

func (tailtracerRcvr *tailtracerReceiver) Start(ctx context.Context, host component.Host) error {
	tailtracerRcvr.host = host
	ctx = context.Background()
	ctx, tailtracerRcvr.cancel = context.WithCancel(ctx)

	interval, _ := time.ParseDuration(tailtracerRcvr.config.Interval)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				tailtracerRcvr.logger.Info("I should start processing traces now!")
			case <-ctx.Done():
				return
			}
		}
	}()

	go tailtracerRcvr.startHTTPServer(ctx)
	return nil
}

func (tailtracerRcvr *tailtracerReceiver) Shutdown(ctx context.Context) error {
	if tailtracerRcvr.cancel != nil {
		tailtracerRcvr.cancel()
	}
	return nil
}

func (tailtracerRcvr *tailtracerReceiver) startHTTPServer(ctx context.Context) {
	http.HandleFunc("/traces", func(w http.ResponseWriter, r *http.Request) {
		file := r.URL.Query().Get("file")
		if file == "" {
			http.Error(w, "Missing 'file' parameter", http.StatusBadRequest)
			return
		}

		tailtracerRcvr.logger.Info("Received a request for traces.", zap.String("file", file))

		f, err := os.Open(file)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer f.Close()

		// Read the file
		for {
			var span Span
			err = binary.Read(f, binary.LittleEndian, &span)
			if err != nil {
				break
			}

			// Convert the name to a string and trim the null bytes
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
			tailtracerRcvr.logger.Fatal("HTTP server ListenAndServe", zap.Error(err))
		}
	}()

	<-ctx.Done()
	if err := server.Shutdown(context.Background()); err != nil {
		tailtracerRcvr.logger.Fatal("HTTP server Shutdown", zap.Error(err))
	}
}
