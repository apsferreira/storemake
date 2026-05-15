// Package pulse — cliente fire-and-forget para o pulse-service.
// Nunca bloqueia o fluxo principal: falhas de tracking são silenciosas.
package pulse

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

const (
	defaultAPIURL     = "https://pulse-api.institutoitinerante.com.br"
	defaultProjectKey = "storemake"
)

// Client envia eventos ao pulse-service de forma assíncrona.
type Client struct {
	apiURL     string
	projectKey string
	httpClient *http.Client
}

type payload struct {
	ProjectKey string            `json:"project_key"`
	EventName  string            `json:"event_name"`
	UserID     string            `json:"user_id,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

// New cria um Client lendo PULSE_API_URL e PULSE_PROJECT_KEY do ambiente.
func New() *Client {
	apiURL := os.Getenv("PULSE_API_URL")
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	projectKey := os.Getenv("PULSE_PROJECT_KEY")
	if projectKey == "" {
		projectKey = defaultProjectKey
	}
	return &Client{
		apiURL:     apiURL,
		projectKey: projectKey,
		httpClient: &http.Client{Timeout: 3 * time.Second},
	}
}

// Track envia um evento ao Pulse de forma assíncrona (fire-and-forget).
func (c *Client) Track(ctx context.Context, eventName, userID string, props map[string]string) {
	go func() {
		body, err := json.Marshal(payload{
			ProjectKey: c.projectKey,
			EventName:  eventName,
			UserID:     userID,
			Properties: props,
		})
		if err != nil {
			return
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			c.apiURL+"/api/v1/events/track", bytes.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return
		}
		resp.Body.Close()
	}()
}
