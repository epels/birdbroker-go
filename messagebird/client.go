package messagebird

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/epels/birdbroker-go"
)

const defaultBaseURL = "https://rest.messagebird.com"

type client struct {
	accessKey, baseURL string
	httpClient         *http.Client
}

// NewClient creates a new MessageBird client with access key ak.
func NewClient(ak string) *client {
	return &client{
		accessKey: ak,
		baseURL:   defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *client) SendMessage(ctx context.Context, m *birdbroker.Message) error {
	data := struct {
		Body       string `json:"body"`
		Originator string `json:"originator"`
		Recipients string `json:"recipients"`
	}{
		Body:       m.Body,
		Originator: m.Originator,
		Recipients: m.Recipient,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encoding/json: Marshal: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/messages", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("net/http: NewRequest: %s", err)
	}
	req.Header.Set("Authorization", "AccessKey "+c.accessKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("net/http: Client.Do: %s", err)
	}
	defer func() {
		// Just close the body: relying on status code for now.
		if err := res.Body.Close(); err != nil {
			log.Printf("%T: Close: %s", res.Body, err)
		}
	}()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("io/ioutil: ReadAll: %s", err)
		}
		return fmt.Errorf("unexpected status code (%d) from MessageBird API with error: %s", res.StatusCode, b)
	}
	return nil
}
