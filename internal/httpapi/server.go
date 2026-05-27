package httpapi

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"hysteria2-web/internal/app"
)

func Start(addr string, a *app.App, logger *slog.Logger) (listenAddr string, err error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("listen %s: %w", addr, err)
	}

	actual := ln.Addr().String()
	handler := NewRouter(a, logger)
	go func() {
		if serveErr := http.Serve(ln, handler); serveErr != nil {
			logger.Error("http server stopped", "addr", actual, "err", serveErr)
		}
	}()
	return actual, nil
}

func ListenAndServe(addr string, a *app.App, logger *slog.Logger) error {
	listenAddr, err := Start(addr, a, logger)
	if err != nil {
		return err
	}
	logger.Info("http server started", "addr", listenAddr)
	select {}
}
