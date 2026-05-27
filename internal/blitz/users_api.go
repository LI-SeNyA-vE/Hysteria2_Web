package blitz

import (
	"context"
	"fmt"
	"net/http"
)

func (c *HTTPClient) ListUsers(ctx context.Context) ([]UserInfo, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/users/", http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("blitz: list users: %w", err)
	}
	var users []UserInfo
	if err := c.decodeJSON(resp, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (c *HTTPClient) AddUser(ctx context.Context, req AddUserRequest) error {
	_, err := c.doDetail(ctx, http.MethodPost, "/api/v1/users/", req, 0)
	if err != nil {
		return err
	}
	return nil
}

func (c *HTTPClient) RemoveUser(ctx context.Context, username string) error {
	path := fmt.Sprintf("/api/v1/users/%s", urlPathEscape(username))
	_, err := c.doDetailNoBody(ctx, http.MethodDelete, path, http.StatusOK)
	if err != nil {
		return fmt.Errorf("blitz: remove user: %w", err)
	}
	return nil
}

func (c *HTTPClient) ShowUserURI(ctx context.Context, username string) (UserURIResponse, error) {
	path := fmt.Sprintf("/api/v1/users/%s/uri", urlPathEscape(username))
	resp, err := c.doNoBody(ctx, http.MethodGet, path, http.StatusOK)
	if err != nil {
		return UserURIResponse{}, fmt.Errorf("blitz: show user uri: %w", err)
	}
	var out UserURIResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return UserURIResponse{}, err
	}
	return out, nil
}
