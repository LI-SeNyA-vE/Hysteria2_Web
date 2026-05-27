package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"hysteria2-web/internal/app"
)

type HTTPServer struct {
	srv *http.Server
	ln  net.Listener
}

func Start(addr string, a *app.App, logger *slog.Logger) (*HTTPServer, string, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", fmt.Errorf("listen %s: %w", addr, err)
	}

	actual := ln.Addr().String()
	hs := &HTTPServer{
		ln: ln,
		srv: &http.Server{
			Handler: NewRouter(a, logger),
		},
	}
	go func() {
		if serveErr := hs.srv.Serve(ln); serveErr != nil && serveErr != http.ErrServerClosed {
			logger.Error("http server stopped", "addr", actual, "err", serveErr)
		}
	}()
	return hs, actual, nil
}

func (hs *HTTPServer) Stop() error {
	if hs == nil || hs.srv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return hs.srv.Shutdown(ctx)
}

func ListenAndServe(addr string, a *app.App, logger *slog.Logger) error {
	hs, listenAddr, err := Start(addr, a, logger)
	if err != nil {
		return err
	}
	logger.Info("http server started", "addr", listenAddr)
	_ = hs
	select {}
}
