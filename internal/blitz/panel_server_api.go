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
