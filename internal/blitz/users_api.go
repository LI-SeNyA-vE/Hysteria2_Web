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
	_, err := c.doDetail(ctx, http.MethodPost, "/api/v1/users/", req, http.StatusCreated)
	if err != nil {
		return fmt.Errorf("blitz: add user: %w", err)
	}
	return nil
}

func (c *HTTPClient) AddBulkUsers(ctx context.Context, req AddBulkUsersRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/users/bulk/", req, http.StatusCreated)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: add bulk users: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) ShowMultipleUserURIs(ctx context.Context, req UsernamesRequest) ([]UserURIResponse, error) {
	resp, err := c.doJSON(ctx, http.MethodPost, "/api/v1/users/uri/bulk", req, http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("blitz: show multiple user uris: %w", err)
	}
	var out []UserURIResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *HTTPClient) BulkRemoveUsers(ctx context.Context, req UsernamesRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/users/bulk-delete", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: bulk remove users: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) GetUser(ctx context.Context, username string) (UserInfo, error) {
	path := fmt.Sprintf("/api/v1/users/%s", urlPathEscape(username))
	resp, err := c.doNoBody(ctx, http.MethodGet, path, http.StatusOK)
	if err != nil {
		return UserInfo{}, fmt.Errorf("blitz: get user: %w", err)
	}
	var out UserInfo
	if err := c.decodeJSON(resp, &out); err != nil {
		return UserInfo{}, err
	}
	return out, nil
}

func (c *HTTPClient) EditUser(ctx context.Context, username string, req EditUserRequest) (DetailResponse, error) {
	path := fmt.Sprintf("/api/v1/users/%s", urlPathEscape(username))
	out, err := c.doDetail(ctx, http.MethodPatch, path, req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: edit user: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) RemoveUser(ctx context.Context, username string) error {
	path := fmt.Sprintf("/api/v1/users/%s", urlPathEscape(username))
	_, err := c.doDetailNoBody(ctx, http.MethodDelete, path, http.StatusOK)
	if err != nil {
		return fmt.Errorf("blitz: remove user: %w", err)
	}
	return nil
}

func (c *HTTPClient) ResetUser(ctx context.Context, username string) (DetailResponse, error) {
	path := fmt.Sprintf("/api/v1/users/%s/reset", urlPathEscape(username))
	out, err := c.doDetailNoBody(ctx, http.MethodGet, path, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: reset user: %w", err)
	}
	return out, nil
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
