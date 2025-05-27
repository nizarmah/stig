// Package agent provides the agent that plays the game.
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nizarmah/stig/game/internal/game"
)

// ClientConfiguration is the configuration for the agent.
type ClientConfiguration struct {
	// APIURL is the URL of the agent API.
	APIURL string
	// Debug is whether to debug the agent client.
	Debug bool
	// Timeout is the timeout for the agent to act.
	Timeout time.Duration
}

// Client is the agent that plays the game.
type Client struct {
	apiURL  string
	debug   bool
	timeout time.Duration
}

// NewClient creates a new client.
func NewClient(cfg ClientConfiguration) *Client {
	return &Client{
		apiURL:  cfg.APIURL,
		debug:   cfg.Debug,
		timeout: cfg.Timeout,
	}
}

// Act returns the action to take on the given frame.
func (c *Client) Act(frame []byte) (game.Action, error) {
	url := fmt.Sprintf("%s/act", c.apiURL)

	// Prepare the request.
	req, _ := http.NewRequest("POST", url, bytes.NewReader(frame))
	req.Header.Set("Content-Type", "image/jpeg")

	// Send the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if c.debug {
			log.Println(fmt.Sprintf("agent failed to send request: %v", err))
		}

		return game.Action{}, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if c.debug {
			log.Println(fmt.Sprintf("agent failed to send request: %v", resp.StatusCode))
		}

		return game.Action{}, fmt.Errorf("failed to send request: %v", resp.StatusCode)
	}

	// Parse the response.
	action := game.Action{}
	if err := json.NewDecoder(resp.Body).Decode(&action); err != nil {
		if c.debug {
			log.Println(fmt.Sprintf("agent failed to decode response: body: %s, err: %v", resp.Body, err))
		}

		return game.Action{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if c.debug {
		log.Println(fmt.Sprintf("agent action: %+v", action))
	}

	return action, nil
}
