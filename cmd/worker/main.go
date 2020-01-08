package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beanstalkd/go-beanstalk"

	"github.com/epels/birdbroker-go"
	"github.com/epels/birdbroker-go/messagebird"
	"github.com/epels/birdbroker-go/queue"
)

type handler struct {
	snd sender
}

type sender interface {
	SendMessage(ctx context.Context, m *birdbroker.Message) error
}

func (c *handler) ServeJob(ctx context.Context, m *birdbroker.Message) error {
	if err := c.snd.SendMessage(ctx, m); err != nil {
		return fmt.Errorf("%T: SendMessage: %s", c.snd, err)
	}
	return nil
}

func main() {
	ak := mustGetenv("MESSAGEBIRD_ACCESS_KEY")
	h := handler{
		snd: messagebird.NewClient(ak),
	}

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
	c := queue.NewConsumer(conn, &h)

	errCh := make(chan error, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting Beanstalk job handler")
		errCh <- c.ListenAndServe()
	}()

	select {
	case err = <-errCh:
		log.Printf("Exiting with error: %s", err)
	case sig := <-sigCh:
		log.Printf("Exiting with signal: %s", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = c.Shutdown(ctx); err != nil {
		log.Fatalf("queue: Consumer.Shutdown: %s", err)
	}
}

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Missing mandatory environment variable %q", key)
	}
	return val
}
