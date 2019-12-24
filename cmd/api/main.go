package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beanstalkd/go-beanstalk"

	"github.com/epels/birdbroker-go/api"
	"github.com/epels/birdbroker-go/queue"
	"github.com/epels/birdbroker-go/service"
)

func main() {
	bsAddr := mustGetenv("BEANSTALK_ADDR")
	conn, err := beanstalk.DialTimeout("tcp", bsAddr, 10*time.Second)
	if err != nil {
		log.Fatalf("beanstalk: DialTimeout: %s", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("beanstalk: Conn.Close: %s", err)
		}
	}()

	mq := queue.NewMessageQueue(conn)
	svc := service.New(mq)
	a := api.NewHandler(svc)

	httpAddr := mustGetenv("HTTP_ADDR")
	s := http.Server{
		Addr:         httpAddr,
		Handler:      a,
		IdleTimeout:  60 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting HTTP server on %q", httpAddr)
		errCh <- s.ListenAndServe()
	}()

	select {
	case err = <-errCh:
		log.Printf("Exiting with error: %s", err)
	case sig := <-sigCh:
		log.Printf("Exiting with signal: %s", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = s.Shutdown(ctx); err != nil {
		log.Fatalf("net/http: Server.Shutdown: %s", err)
	}
}

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Missing mandatory environment variable %q", key)
	}
	return val
}
