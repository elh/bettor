package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	"github.com/elh/bettor/api/bettor/v1alpha/bettorv1alphaconnect"
	"github.com/elh/bettor/internal/app/bettor/discord"
	"github.com/elh/bettor/internal/app/bettor/repo/gob"
	"github.com/elh/bettor/internal/app/bettor/server"
	"github.com/go-kit/log"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// TODO: have Makefile commands run with environment variables -> flags.
var (
	port         = flag.Int("port", 8080, "The server port")
	discordToken = flag.String("discordToken", "", "Discord bot token (secret)")
)

const (
	serviceName = "bettor"
	gobDBFile   = "bettor.gob"
)

func init() {
	flag.Parse()
}

func main() {
	logger := log.NewJSONLogger(os.Stdout)

	// server with gob file-backed repo
	r, err := gob.New(gobDBFile)
	if err != nil {
		logger.Log("msg", "error creating repo", "err", err)
		panic(err)
	}
	s := server.New(r)

	// tracing
	tp, err := tracerProvider(os.Stdout)
	if err != nil {
		logger.Log("msg", "error creating tracer provider", "err", err)
		panic(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Log("msg", "error shutting down tracer provider", "err", err)
		}
	}()

	// exit if either goroutine exits
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Buf connect server
	go func() {
		defer wg.Done()
		mux := http.NewServeMux()
		path, handler := bettorv1alphaconnect.NewBettorServiceHandler(s, otelconnect.WithTelemetry(otelconnect.WithTracerProvider(tp)))
		mux.Handle(path, handler)
		httpServer := http.Server{
			Addr:              fmt.Sprintf("localhost:%d", *port),
			Handler:           h2c.NewHandler(mux, &http2.Server{}),
			ReadHeaderTimeout: 2 * time.Second,
		}
		if err := httpServer.ListenAndServe(); err != nil {
			logger.Log("msg", "http server error", "err", err)
			panic(err)
		}
	}()

	// Discord bot
	go func() {
		defer wg.Done()
		// TODO: use a real client for Bettor service so we get telemetry
		bot, err := discord.New(*discordToken, s, logger)
		if err != nil {
			logger.Log("msg", "error creating discord bot", "err", err)
			panic(err)
		}
		if err := bot.Run(); err != nil {
			logger.Log("msg", "discord bot run exited", "err", err)
			panic(err)
		}
	}()

	// TODO: shutdown is sloppy. we want graceful clean up in the bot but we should be taking the os kill signals here up top
	wg.Wait()
	logger.Log("msg", "exiting")
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
