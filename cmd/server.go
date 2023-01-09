package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/elh/bettor/api/bettor/v1alpha/bettorv1alphaconnect"
	"github.com/elh/bettor/internal/app/bettor/server"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var port = flag.Int("port", 8080, "The server port")

func main() {
	flag.Parse()

	s := server.New()

	mux := http.NewServeMux()
	path, handler := bettorv1alphaconnect.NewBettorServiceHandler(s)
	mux.Handle(path, handler)
	httpServer := http.Server{
		Addr:              fmt.Sprintf("localhost:%d", *port),
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: 2 * time.Second,
	}
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
