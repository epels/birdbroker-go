package queue

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/epels/birdbroker-go"
)

var ErrConsumerClosed = errors.New("consumer was closed")

type consumer struct {
	conn consumerConn
	h    handler

	stopCh chan struct{}
}

type consumerConn interface {
	// Bury sets a job to the "buried" state so it will not be picked up from
	// the queue again. This state is intended for jobs that are considered
	// "faulty", and need to be inspected manually by a human.
	Bury(id uint64, pri uint32) error
	// Delete drops a job from the queue. This should be the final state for
	// any jobs that are handled successfully, or for well-formed jobs that
	// should not be processed now, nor in the future.
	Delete(id uint64) error
	// Release releases a reserved job so it can be picked up by the worker
	// again, much like a retry.
	Release(id uint64, pri uint32, delay time.Duration) error
	// Reserve retrieves a job from the queue and marks it as reserved.
	Reserve(timeout time.Duration) (id uint64, body []byte, err error)
}

type handler interface {
	ServeJob(ctx context.Context, m *birdbroker.Message) error
}

type handlerFunc func(ctx context.Context, m *birdbroker.Message) error

func (f handlerFunc) ServeJob(ctx context.Context, m *birdbroker.Message) error {
	return f(ctx, m)
}

func NewConsumer(c consumerConn, h handler) *consumer {
	return &consumer{
		stopCh: make(chan struct{}, 1),
		conn:   c,
		h:      h,
	}
}

// ListenAndServe consumes jobs from c.producerConn and then calls
// Serve to handle these.
//
// ListenAndServe always returns a non-nil error. After Shutdown or Close, the
// returned error is ErrServerClosed.
func (c *consumer) ListenAndServe() error {
	for {
		select {
		case <-c.stopCh:
			return ErrConsumerClosed
		default:
			// @todo: Make timeout configurable. For now: very long, as we're
			//        operating in a worker context.
			id, b, err := c.conn.Reserve(42 * time.Hour)

			go func(id uint64, b []byte, err error) {
				if err != nil {
					log.Printf("%T: Reserve: %s", c.conn, err)
					return
				}

				var m birdbroker.Message
				if err := json.Unmarshal(b, &m); err != nil {
					log.Printf("encoding/json: Unmarshal: %s", err)

					// Bury the job: its payload has an invalid format, so
					// there's no use in retrying, but it makes sense to
					// inspect the job manually.
					if err = c.conn.Bury(id, defaultPriority); err != nil {
						log.Printf("%T: Delete: %s", c.h, err)
					}
					return
				}

				// Create a fresh context for the handler: replicate net/http
				// behaviour.
				if err := c.h.ServeJob(context.Background(), &m); err != nil {
					log.Printf("%T: ServeJob: %s", c.h, err)
					if err = c.conn.Release(id, defaultPriority, 0); err != nil {
						log.Printf("%T: Release: %s", c.h, err)
					}
					return
				}

				if err = c.conn.Delete(id); err != nil {
					log.Printf("%T: Delete: %s", c.h, err)
				}
			}(id, b, err)
		}
	}
}

// @todo: Graceful shutdown. Accept ctx for now, so I can do so later without
//        making breaking changes to the exported API.
func (c *consumer) Shutdown(ctx context.Context) error {
	c.stopCh <- struct{}{}
	return nil
}
