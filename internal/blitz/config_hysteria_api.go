package blitz

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func (c *HTTPClient) UpdateHysteria(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodPatch, "/api/v1/config/hysteria/update", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: update hysteria: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) RestartHysteria(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodPost, "/api/v1/config/hysteria/restart", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: restart hysteria: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) GetHysteriaPort(ctx context.Context) (GetPortResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/hysteria/get-port", http.StatusOK)
	if err != nil {
		return GetPortResponse{}, fmt.Errorf("blitz: get hysteria port: %w", err)
	}
	var out GetPortResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return GetPortResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) SetHysteriaPort(ctx context.Context, port int) (DetailResponse, error) {
	path := fmt.Sprintf("/api/v1/config/hysteria/set-port/%d", port)
	out, err := c.doDetailNoBody(ctx, http.MethodGet, path, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: set hysteria port: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) GetHysteriaSNI(ctx context.Context) (GetSNIResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/hysteria/get-sni", http.StatusOK)
	if err != nil {
		return GetSNIResponse{}, fmt.Errorf("blitz: get hysteria sni: %w", err)
	}
	var out GetSNIResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return GetSNIResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) SetHysteriaSNI(ctx context.Context, sni string) (DetailResponse, error) {
	path := fmt.Sprintf("/api/v1/config/hysteria/set-sni/%s", urlPathEscape(sni))
	out, err := c.doDetailNoBody(ctx, http.MethodGet, path, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: set hysteria sni: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) BackupHysteriaConfig(ctx context.Context) ([]byte, error) {
	data, err := c.doBytes(ctx, http.MethodGet, "/api/v1/config/hysteria/backup", http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("blitz: backup hysteria config: %w", err)
	}
	return data, nil
}

func (c *HTTPClient) RestoreHysteriaConfig(ctx context.Context, filename string, file io.Reader) (DetailResponse, error) {
	out, err := c.doMultipart(ctx, "/api/v1/config/hysteria/restore", "file", filename, file, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: restore hysteria config: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) EnableHysteriaObfs(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodGet, "/api/v1/config/hysteria/enable-obfs", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: enable hysteria obfs: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) DisableHysteriaObfs(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodGet, "/api/v1/config/hysteria/disable-obfs", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: disable hysteria obfs: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) CheckHysteriaObfs(ctx context.Context) (GetObfsResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/hysteria/check-obfs", http.StatusOK)
	if err != nil {
		return GetObfsResponse{}, fmt.Errorf("blitz: check hysteria obfs: %w", err)
	}
	var out GetObfsResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return GetObfsResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) EnableHysteriaMasquerade(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodGet, "/api/v1/config/hysteria/enable-masquerade", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: enable hysteria masquerade: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) DisableHysteriaMasquerade(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodGet, "/api/v1/config/hysteria/disable-masquerade", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: disable hysteria masquerade: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) CheckHysteriaMasquerade(ctx context.Context) (GetMasqueradeStatusResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/hysteria/check-masquerade", http.StatusOK)
	if err != nil {
		return GetMasqueradeStatusResponse{}, fmt.Errorf("blitz: check hysteria masquerade: %w", err)
	}
	var out GetMasqueradeStatusResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return GetMasqueradeStatusResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) GetHysteriaConfigFile(ctx context.Context) (ConfigFile, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/hysteria/file", http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("blitz: get hysteria config file: %w", err)
	}
	var out ConfigFile
	if err := c.decodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *HTTPClient) SetHysteriaConfigFile(ctx context.Context, file ConfigFile) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/hysteria/file", file, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: set hysteria config file: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) StartIPLimit(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodPost, "/api/v1/config/hysteria/ip-limit/start", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: start ip limit: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) StopIPLimit(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodPost, "/api/v1/config/hysteria/ip-limit/stop", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: stop ip limit: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) CleanIPLimit(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodPost, "/api/v1/config/hysteria/ip-limit/clean", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: clean ip limit: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) GetIPLimitConfig(ctx context.Context) (IPLimitConfigResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/hysteria/ip-limit/config", http.StatusOK)
	if err != nil {
		return IPLimitConfigResponse{}, fmt.Errorf("blitz: get ip limit config: %w", err)
	}
	var out IPLimitConfigResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return IPLimitConfigResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) SetIPLimitConfig(ctx context.Context, cfg IPLimitConfig) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/hysteria/ip-limit/config", cfg, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: set ip limit config: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) SetupWebPanelDecoy(ctx context.Context, req SetupDecoyRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/hysteria/webpanel/decoy/setup", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: setup webpanel decoy: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) StopWebPanelDecoy(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodPost, "/api/v1/config/hysteria/webpanel/decoy/stop", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: stop webpanel decoy: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) GetWebPanelDecoyStatus(ctx context.Context) (DecoyStatusResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/hysteria/webpanel/decoy/status", http.StatusOK)
	if err != nil {
		return DecoyStatusResponse{}, fmt.Errorf("blitz: get webpanel decoy status: %w", err)
	}
	var out DecoyStatusResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return DecoyStatusResponse{}, err
	}
	return out, nil
}
