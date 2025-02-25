package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	hacNodeBaseURL = "http://localhost:8633/api"
	timeout        = 5 * time.Second
)

// HACNodeClient encapsulates communication with HAC node
type HACNodeClient struct {
	client    *http.Client
	authToken string
	baseURL   string
}

func NewHACNodeClient(authToken string) *HACNodeClient {
	return &HACNodeClient{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 5,
			},
		},
		authToken: authToken,
		baseURL:   hacNodeBaseURL,
	}
}

// SetBaseURL allows changing the base URL
func (c *HACNodeClient) SetBaseURL(url string) {
	c.baseURL = url
}

// Register registers service to the service registry via HTTP POST
func (c *HACNodeClient) Register(ctx context.Context, name string, agentUrl string, selfIntro string) error {
	registrationReq := struct {
		Name      string `json:"name"`
		AgentUrl  string `json:"agentUrl"`
		SelfIntro string `json:"selfIntro"`
	}{
		Name:      name,
		AgentUrl:  agentUrl,
		SelfIntro: selfIntro,
	}

	reqBody, err := json.Marshal(registrationReq)
	if err != nil {
		return fmt.Errorf("marshal registration request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/register-agent", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("create registration request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed with status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// PostProposal sends proposal to HAC node
func (c *HACNodeClient) PostProposal(ctx context.Context, data string, title string) error {
	proposalReq := struct {
		Data  string `json:"data"`
		Title string `json:"title"`
	}{
		Data:  data,
		Title: title,
	}

	reqBody, err := json.Marshal(proposalReq)
	if err != nil {
		return fmt.Errorf("marshal proposal request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/post-pr", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("create proposal request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("proposal request failed: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("proposal failed: %s", response.Error)
	}

	log.Printf("Successfully posted proposal: %s", title)
	return nil
}
