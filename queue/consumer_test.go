package queue

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/epels/birdbroker-go"
	"github.com/epels/birdbroker-go/internal/mock"
)

func TestListenAndServe(t *testing.T) {
	t.Run("Closed", func(t *testing.T) {
		hf := handlerFunc(func(ctx context.Context, m *birdbroker.Message) error {
			t.Fatalf("Must never be called")
			return nil
		})
		cons := NewConsumer(nil, hf)

		if err := cons.Shutdown(context.Background()); err != nil {
			t.Errorf("Consumer: Shutdown: %s", err)
		}
		if err := cons.ListenAndServe(); !errors.Is(err, ErrConsumerClosed) {
			t.Errorf("Got %T (%s), expected ErrConsumerClosed", err, err)
		}
	})

	t.Run("Retries failed jobs", func(t *testing.T) {
		// once is used to only return a single job from Reserve.
		var once sync.Once
		// wg is decremented within the ReleaseFunc, so we can wait for it to
		// be invoked before shutting down the consumer later (as
		// ListenAndServe is ran on its own goroutine).
		var wg sync.WaitGroup

		var handlerCalled bool
		hf := handlerFunc(func(ctx context.Context, m *birdbroker.Message) error {
			handlerCalled = true

			if m.Body != "Hello" {
				t.Errorf("Got %q, expected Hello", m.Body)
			}
			if m.Originator != "Foo Inc" {
				t.Errorf("Got %q, expected Foo Inc", m.Originator)
			}
			if m.Recipient != "31612345678" {
				t.Errorf("Got %q, expected 31612345678", m.Recipient)
			}

			return errors.New("some error, so that the job is released")
		})
		var released bool
		cons := NewConsumer(&mock.ConsumerConn{
			ReleaseFunc: func(id uint64, pri uint32, delay time.Duration) error {
				defer wg.Done()
				released = true

				if id != 42 {
					t.Errorf("Got %d, expected 42", id)
				}
				return nil
			},
			ReserveFunc: func(timeout time.Duration) (id uint64, body []byte, err error) {
				// Only return a job on the first invocation.
				once.Do(func() {
					b := []byte(`{
	"body": "Hello",
	"originator": "Foo Inc",
	"recipient": "31612345678"
}`)
					id = uint64(42)
					body = b
					err = nil
				})

				if body == nil {
					time.Sleep(timeout)
					err = errors.New("timeout")
				}
				return
			},
		}, hf)

		wg.Add(1)
		go func() {
			// Run on separate goroutine: ListenAndServe blocks until it's
			// closed.
			if err := cons.ListenAndServe(); !errors.Is(err, ErrConsumerClosed) {
				t.Errorf("Got %T (%s), expected ErrConsumerClosed", err, err)
			}
		}()

		wg.Wait()
		if err := cons.Shutdown(context.Background()); err != nil {
			t.Errorf("Consumer: Shutdown: %s", err)
		}

		if !handlerCalled {
			t.Errorf("Got false, expected true")
		}
		if !released {
			t.Errorf("Got false, expected true")
		}
	})

	t.Run("Deletes successful jobs", func(t *testing.T) {
		// once is used to only return a single job from Reserve.
		var once sync.Once
		// wg is decremented within the DeleteFunc, so we can wait for it to be
		// invoked before shutting down the consumer later (as ListenAndServe
		// is ran on its own goroutine).
		var wg sync.WaitGroup

		var handlerCalled bool
		hf := handlerFunc(func(ctx context.Context, m *birdbroker.Message) error {
			handlerCalled = true

			if m.Body != "Hello" {
				t.Errorf("Got %q, expected Hello", m.Body)
			}
			if m.Originator != "Foo Inc" {
				t.Errorf("Got %q, expected Foo Inc", m.Originator)
			}
			if m.Recipient != "31612345678" {
				t.Errorf("Got %q, expected 31612345678", m.Recipient)
			}

			return nil
		})
		var deleted bool
		cons := NewConsumer(&mock.ConsumerConn{
			DeleteFunc: func(id uint64) error {
				defer wg.Done()
				deleted = true

				if id != 42 {
					t.Errorf("Got %d, expected 42", id)
				}
				return nil
			},
			ReserveFunc: func(timeout time.Duration) (id uint64, body []byte, err error) {
				// Only return a job on the first invocation.
				once.Do(func() {
					b := []byte(`{
	"body": "Hello",
	"originator": "Foo Inc",
	"recipient": "31612345678"
}`)
					id = uint64(42)
					body = b
					err = nil
				})

				if body == nil {
					time.Sleep(timeout)
					err = errors.New("timeout")
				}
				return
			},
		}, hf)

		wg.Add(1)
		go func() {
			// Run on separate goroutine: ListenAndServe blocks until it's
			// closed.
			if err := cons.ListenAndServe(); !errors.Is(err, ErrConsumerClosed) {
				t.Errorf("Got %T (%s), expected ErrConsumerClosed", err, err)
			}
		}()

		wg.Wait()
		if err := cons.Shutdown(context.Background()); err != nil {
			t.Errorf("Consumer: Shutdown: %s", err)
		}

		if !handlerCalled {
			t.Errorf("Got false, expected true")
		}
		if !deleted {
			t.Errorf("Got false, expected true")
		}
	})
}
