package blitz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type UserClient interface {
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

func (c *HTTPClient) BaseURL() string {
	return c.baseURL
}

func (c *HTTPClient) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Accept", "application/json")
}

func (c *HTTPClient) setJSONHeaders(req *http.Request) {
	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")
}

func (c *HTTPClient) url(path string) string {
	return c.baseURL + path
}

func (c *HTTPClient) do(ctx context.Context, req *http.Request, expectStatus int) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("blitz: HTTP %s %s: %w", req.Method, req.URL.String(), err)
	}
	if !statusMatches(resp.StatusCode, expectStatus) {
		defer resp.Body.Close()
		return nil, c.parseError(resp, req.Method+" "+req.URL.Path)
	}
	return resp, nil
}

// expectStatus: exact code, or 0 to accept any 2xx.
func statusMatches(code, expectStatus int) bool {
	if expectStatus == 0 {
		return code >= 200 && code < 300
	}
	return code == expectStatus
}

func (c *HTTPClient) doJSON(ctx context.Context, method, path string, body any, expectStatus int) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("blitz: marshal request: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.url(path), reader)
	if err != nil {
		return nil, fmt.Errorf("blitz: create request: %w", err)
	}
	c.setJSONHeaders(req)
	return c.do(ctx, req, expectStatus)
}

func (c *HTTPClient) doNoBody(ctx context.Context, method, path string, expectStatus int) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.url(path), nil)
	if err != nil {
		return nil, fmt.Errorf("blitz: create request: %w", err)
	}
	c.setJSONHeaders(req)
	return c.do(ctx, req, expectStatus)
}

func (c *HTTPClient) decodeJSON(resp *http.Response, dst any) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("blitz: read response: %w", err)
	}
	if dst == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("blitz: decode response: %w (body: %s)", err, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *HTTPClient) doDetail(ctx context.Context, method, path string, body any, expectStatus int) (DetailResponse, error) {
	resp, err := c.doJSON(ctx, method, path, body, expectStatus)
	if err != nil {
		return DetailResponse{}, err
	}
	var out DetailResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return DetailResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) doDetailNoBody(ctx context.Context, method, path string, expectStatus int) (DetailResponse, error) {
	resp, err := c.doNoBody(ctx, method, path, expectStatus)
	if err != nil {
		return DetailResponse{}, err
	}
	var out DetailResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return DetailResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) doBytes(ctx context.Context, method, path string, expectStatus int) ([]byte, error) {
	resp, err := c.doNoBody(ctx, method, path, expectStatus)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *HTTPClient) doMultipart(ctx context.Context, path, fieldName, filename string, file io.Reader, expectStatus int) (DetailResponse, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: write form file: %w", err)
	}
	if err := writer.Close(); err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(path), &buf)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: create request: %w", err)
	}
	c.setHeaders(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.do(ctx, req, expectStatus)
	if err != nil {
		return DetailResponse{}, err
	}
	var out DetailResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return DetailResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) parseError(resp *http.Response, action string) error {
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

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
