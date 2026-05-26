package blitz

import (
	"context"
	"fmt"
	"net/http"
)

func (c *HTTPClient) InstallWarp(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodPost, "/api/v1/config/warp/install", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: install warp: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) UninstallWarp(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodDelete, "/api/v1/config/warp/uninstall", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: uninstall warp: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) ConfigureWarp(ctx context.Context, req ConfigureWarpRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/warp/configure", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: configure warp: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) GetWarpStatus(ctx context.Context) (WarpStatusResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/warp/status", http.StatusOK)
	if err != nil {
		return WarpStatusResponse{}, fmt.Errorf("blitz: get warp status: %w", err)
	}
	var out WarpStatusResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return WarpStatusResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) StartTelegramBot(ctx context.Context, req TelegramStartRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/telegram/start", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: start telegram bot: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) StopTelegramBot(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodDelete, "/api/v1/config/telegram/stop", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: stop telegram bot: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) GetTelegramBackupInterval(ctx context.Context) (BackupIntervalResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/telegram/backup-interval", http.StatusOK)
	if err != nil {
		return BackupIntervalResponse{}, fmt.Errorf("blitz: get telegram backup interval: %w", err)
	}
	var out BackupIntervalResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return BackupIntervalResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) SetTelegramBackupInterval(ctx context.Context, req SetIntervalRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/telegram/backup-interval", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: set telegram backup interval: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) StartNormalSub(ctx context.Context, req DomainPortRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/normalsub/start", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: start normalsub: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) StopNormalSub(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodDelete, "/api/v1/config/normalsub/stop", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: stop normalsub: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) EditNormalSubSubpath(ctx context.Context, req EditSubPathRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPut, "/api/v1/config/normalsub/edit_subpath", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: edit normalsub subpath: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) GetNormalSubSubpath(ctx context.Context) (GetSubPathResponse, error) {
	resp, err := c.doNoBody(ctx, http.MethodGet, "/api/v1/config/normalsub/subpath", http.StatusOK)
	if err != nil {
		return GetSubPathResponse{}, fmt.Errorf("blitz: get normalsub subpath: %w", err)
	}
	var out GetSubPathResponse
	if err := c.decodeJSON(resp, &out); err != nil {
		return GetSubPathResponse{}, err
	}
	return out, nil
}

func (c *HTTPClient) StartSingbox(ctx context.Context, req DomainPortRequest) (DetailResponse, error) {
	out, err := c.doDetail(ctx, http.MethodPost, "/api/v1/config/singbox/start", req, http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: start singbox: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) StopSingbox(ctx context.Context) (DetailResponse, error) {
	out, err := c.doDetailNoBody(ctx, http.MethodDelete, "/api/v1/config/singbox/stop", http.StatusOK)
	if err != nil {
		return DetailResponse{}, fmt.Errorf("blitz: stop singbox: %w", err)
	}
	return out, nil
}
