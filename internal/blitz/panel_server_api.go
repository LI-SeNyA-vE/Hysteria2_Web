package blitz

import (
	"context"
	"fmt"
	"net/http"
)

func (c *HTTPClient) GetServerStatus(ctx context.Context) (ServerStatusResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/server/status", http.StatusOK)
	if err != nil {
		return ServerStatusResponse{}, fmt.Errorf("blitz: get server status: %w", err)
	}
	var out ServerStatusResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return ServerStatusResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) GetServerServicesStatus(ctx context.Context) (ServerServicesStatusResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/server/services/status", http.StatusOK)
	if err != nil {
		return ServerServicesStatusResponse{}, fmt.Errorf("blitz: get server services status: %w", err)
	}
	var out ServerServicesStatusResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return ServerServicesStatusResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) GetVersionInfo(ctx context.Context) (VersionInfoResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/server/version", http.StatusOK)
	if err != nil {
		return VersionInfoResponse{}, fmt.Errorf("blitz: get version info: %w", err)
	}
	var out VersionInfoResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return VersionInfoResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) CheckVersionInfo(ctx context.Context) (VersionCheckResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/server/version/check", http.StatusOK)
	if err != nil {
		return VersionCheckResponse{}, fmt.Errorf("blitz: check version info: %w", err)
	}
	var out VersionCheckResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return VersionCheckResponse{}, err
	}
	return out, nil
}
