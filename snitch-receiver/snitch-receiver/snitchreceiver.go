package snitchreceiver

/*
   #include "data.h"
*/
import "C"
import (
	"context"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"unsafe"

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

func readFile(file string) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	// map f to memory through mmap
	mmap, err := syscall.Mmap(int(f.Fd()), 0, C.MAX_FILE_SIZE, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		fmt.Println("Error mapping file:", err)
		return
	}

	fmt.Println("mmap is", mmap[:1000])
	spanOffsetSlice := *(*[]int64)(unsafe.Pointer(&mmap))
	fmt.Println("spanOffsetSlice is", spanOffsetSlice[:100])

	for i, spanOffset := range spanOffsetSlice[:C.MAX_SPANS] {
		if spanOffset == 0 {
			continue
		}

		fmt.Printf("xxx i: %d, spanOffset: %d\n", i, spanOffset)

		span := (*C.Span)(unsafe.Pointer(&mmap[spanOffset]))
		fmt.Printf("Span kind is %d\n", span.kind)
		nameOffset := uintptr(span.name_offset)
		name := C.GoString((*C.char)(unsafe.Pointer(&mmap[nameOffset])))
		fmt.Printf("Name: %s\n", name)

		// print status
		status := (*C.Status)(unsafe.Pointer(&mmap[span.status_offset]))
		fmt.Printf("Status offset: %d, code: %d\n", span.status_offset, status.code)

		// parent_span_id
		parent_span_id := C.GoString((*C.char)(unsafe.Pointer(&mmap[span.parent_span_id_offset])))
		fmt.Printf("Parent span id: %s\n", parent_span_id)

		// start_time
		start_time := C.GoString((*C.char)(unsafe.Pointer(&mmap[span.start_timestamp_offset])))
		fmt.Printf("Start time: %s\n", start_time)

		// end_time
		end_time := C.GoString((*C.char)(unsafe.Pointer(&mmap[span.end_timestamp_offset])))
		fmt.Printf("End time: %s\n", end_time)

		// print attributes
		attributes_count := span.total_recorded_attributes
		fmt.Printf("attributes_count: %d\n", attributes_count)
		keyValues := (*[100]C.KeyValue)(unsafe.Pointer(&mmap[span.attributes_offset]))[:attributes_count:attributes_count]

		for j, kv := range keyValues {
			fmt.Printf("x============ %d\n", j)
			value := (*C.AnyValue)(unsafe.Pointer(&mmap[kv.value_offset]))

			switch value.value_type {
			case C.ANYVALUE_STRING:
				stringOffset := *(*C.size_t)(unsafe.Pointer(&value.value))
				char := C.GoString((*C.char)(unsafe.Pointer(&mmap[stringOffset])))
				fmt.Printf("attribute value: %s\n", char)
			case C.ANYVALUE_INT:
				fmt.Println("Value type is INT ..............")
			default:
				fmt.Println("Unknown value type or not implemented")
			}

			fmt.Printf("KeyValue kv: %v, value: %v\n", kv, value)

			key := C.GoString((*C.char)(unsafe.Pointer(&mmap[kv.key_offset])))
			fmt.Printf("KeyValue %d: key=%s\n", j, key)
		}
	}
}

func (snitchReceiver *SnitchReceiver) startHTTPServer(ctx context.Context) {
	http.HandleFunc("/traces", func(w http.ResponseWriter, r *http.Request) {
		file := r.URL.Query().Get("file")
		if file == "" {
			http.Error(w, "Missing 'file' parameter", http.StatusBadRequest)
			return
		}

		snitchReceiver.logger.Info("Received a request for traces.", zap.String("file", file))
		readFile(file)

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
