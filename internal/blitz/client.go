package blitz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client interface {
	AddUser(ctx context.Context, req AddUserRequest) error
	RemoveUser(ctx context.Context, username string) error
	ListUsers(ctx context.Context) ([]UserInfo, error)
}

type HTTPClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *HTTPClient) AddUser(ctx context.Context, req AddUserRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("blitz: marshal add user request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/users/", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("blitz: create add user request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("blitz: add user request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return nil
	}
	return c.parseError(resp, "add user")
}

func (c *HTTPClient) RemoveUser(ctx context.Context, username string) error {
	path := fmt.Sprintf("%s/api/v1/users/%s", c.baseURL, urlPathEscape(username))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("blitz: create remove user request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("blitz: remove user request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	return c.parseError(resp, "remove user")
}

func (c *HTTPClient) ListUsers(ctx context.Context) ([]UserInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/users/", nil)
	if err != nil {
		return nil, fmt.Errorf("blitz: create list users request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("blitz: list users request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp, "list users")
	}

	var users []UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("blitz: decode list users response: %w", err)
	}
	return users, nil
}

func (c *HTTPClient) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
}

func (c *HTTPClient) parseError(resp *http.Response, action string) error {
	body, _ := io.ReadAll(resp.Body)

	var detail DetailResponse
	if err := json.Unmarshal(body, &detail); err == nil && detail.Detail != "" {
		return fmt.Errorf("blitz: %s: status %d: %s", action, resp.StatusCode, detail.Detail)
	}

	var validation HTTPValidationError
	if err := json.Unmarshal(body, &validation); err == nil && len(validation.Detail) > 0 {
		return fmt.Errorf("blitz: %s: status %d: %s", action, resp.StatusCode, validation.Detail[0].Msg)
	}

	return fmt.Errorf("blitz: %s: status %d: %s", action, resp.StatusCode, strings.TrimSpace(string(body)))
}

func urlPathEscape(s string) string {
	return strings.ReplaceAll(s, "/", "%2F")
}
