package blitz

import (
	"context"
	"fmt"
	"net/http"
)

func (c *HTTPClient) GetIPStatus(ctx context.Context) (IPStatusResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/ip/get", http.StatusOK)
	if err != nil {
		return IPStatusResponse{}, fmt.Errorf("blitz: get ip status: %w", err)
	}
	var out IPStatusResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return IPStatusResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) AddIP(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodGet, "/api/v1/config/ip/add", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: add ip: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) EditIP(ctx context.Context, req EditIPRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/ip/edit", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: edit ip: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) GetAllNodes(ctx context.Context) ([]Node, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/ip/nodes", http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("blitz: get all nodes: %w", err)
	}
	var out []Node
	if err := c.decodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *HTTPClient) AddNode(ctx context.Context, req AddNodeRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/ip/nodes/add", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: add node: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) DeleteNode(ctx context.Context, req DeleteNodeRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/ip/nodes/delete", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: delete node: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) ReceiveNodeTraffic(ctx context.Context, req NodesTrafficPayload) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/ip/nodestraffic", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: receive node traffic: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) GetAllExtraConfigs(ctx context.Context) ([]ExtraConfigResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/extra-config/list", http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("blitz: get all extra configs: %w", err)
	}
	var out []ExtraConfigResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *HTTPClient) AddExtraConfig(ctx context.Context, req AddExtraConfigRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/extra-config/add", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: add extra config: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) DeleteExtraConfig(ctx context.Context, req DeleteExtraConfigRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/extra-config/delete", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: delete extra config: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) InstallTCPBrutal(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodPost, "/api/v1/config/install-tcp-brutal", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: install tcp brutal: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) UpdateGeo(ctx context.Context, country string) (DetailResponse, error) {
	path := fmt.Sprintf("/api/v1/config/update-geo/%s", urlPathEscape(country))
	out, err := c.doDetailNoBody(ctx, http.MethodGet, path, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: update geo: %w", err)
	}
	return out, nil
}
