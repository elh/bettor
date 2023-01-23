package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	"github.com/elh/bettor/api/bettor/v1alpha/bettorv1alphaconnect"
	"github.com/elh/bettor/internal/app/bettor/discord"
	"github.com/elh/bettor/internal/app/bettor/repo/gob"
	"github.com/elh/bettor/internal/app/bettor/server"
	"github.com/elh/bettor/internal/pkg/envflag"
	"github.com/go-kit/log"
	_ "github.com/joho/godotenv/autoload" // loads .env file before envflag reads from them
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	port         = envflag.Int("port", 8080, "The server port")
	discordToken = envflag.String("discordToken", "", "Discord bot token (secret)")
	gobDBFile    = envflag.String("gobDBFile", "bettor.gob", "Gob file to use for persistence")
)

const (
	serviceName = "bettor"
)

func init() {
	envflag.Parse()
}

func main() {
	logger := log.NewJSONLogger(os.Stdout)
	serverLogger := log.With(logger, "component", "server")

	// Server with gob file-backed repo
	r, err := gob.New(*gobDBFile)
	if err != nil {
		logger.Log("msg", "error creating repo", "err", err)
		panic(err)
	}
	s, err := server.New(server.WithRepo(r), server.WithLogger(serverLogger))
	if err != nil {
		logger.Log("msg", "error creating server", "err", err)
		panic(err)
	}

	// Tracing
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

	// Use context cancellation to coordinate graceful shutdown. Exit on interrupt or any worker exiting.
	ctx, cancelFn := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Buf Connect server
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
		go func() {
			if err := httpServer.ListenAndServe(); err != nil {
				serverLogger.Log("msg", "http server error", "err", err)
				cancelFn()
				return
			}
		}()
		<-ctx.Done()
		if err := httpServer.Shutdown(ctx); err != nil {
			serverLogger.Log("msg", "error shutting down http server", "err", err)
		}
	}()

	// Discord bot
	go func() {
		defer wg.Done()
		botLogger := log.With(logger, "component", "discord-bot")
		netClient := &http.Client{
			Timeout: time.Second * 5,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		}
		client := bettorv1alphaconnect.NewBettorServiceClient(netClient, fmt.Sprintf("http://localhost:%d", *port))
		bot, err := discord.New(ctx, discord.WithToken(*discordToken), discord.WithBettorClient(client), discord.WithLogger(botLogger))
		if err != nil {
			botLogger.Log("msg", "error creating discord bot", "err", err)
			cancelFn()
			return
		}
		if err := bot.Run(); err != nil {
			botLogger.Log("msg", "discord bot run exited", "err", err)
			cancelFn()
			return
		}
	}()

	// Wait for graceful shutdown of server and discord bot
	wg.Wait()
	logger.Log("msg", "exiting")
}

func tracerProvider(w io.Writer) (*trace.TracerProvider, error) {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("component", "server"),
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
