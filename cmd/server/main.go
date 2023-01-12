package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	"github.com/elh/bettor/api/bettor/v1alpha/bettorv1alphaconnect"
	"github.com/elh/bettor/internal/app/bettor/repo/gob"
	"github.com/elh/bettor/internal/app/bettor/server"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var port = flag.Int("port", 8080, "The server port")

const serviceName = "bettor"

const gobDBFile = "bettor.gob"

func main() {
	flag.Parse()

	// server with gob file-backed repo
	r, err := gob.New(gobDBFile)
	if err != nil {
		log.Fatal(err)
	}
	s := server.New(r)

	// tracing
	tp, err := tracerProvider(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	// Buf connect server
	mux := http.NewServeMux()
	path, handler := bettorv1alphaconnect.NewBettorServiceHandler(s, otelconnect.WithTelemetry(otelconnect.WithTracerProvider(tp)))
	mux.Handle(path, handler)
	httpServer := http.Server{
		Addr:              fmt.Sprintf("localhost:%d", *port),
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: 2 * time.Second,
	}
	if err := httpServer.ListenAndServe(); err != nil {
		log.Default().Print(err)
		return
	}
}

func tracerProvider(w io.Writer) (*trace.TracerProvider, error) {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	e, err := stdouttrace.New(
		stdouttrace.WithWriter(w),
	)
	if err != nil {
		return nil, err
	}

	return trace.NewTracerProvider(
		trace.WithBatcher(e),
		trace.WithResource(r),
	), nil
}
