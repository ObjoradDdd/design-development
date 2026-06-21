package main

import (
	"context"
	"design-developer/internal/cache"
	"design-developer/internal/compressor"
	"design-developer/internal/config"
	"design-developer/internal/logs"
	"design-developer/internal/middleware"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {

	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	logs.ConfigureSlogLogging(config.LogLevel)

	slog.Info("Starting reverse proxy server...")

	zs, err := compressor.NewZstdCompressor(config.ZstdLevel)
	if err != nil {
		slog.Error("failed to create Zstd compressor", "error", err)
		return
	}

	gz := compressor.NewGzipCompressor(config.GzipLevel)

	targetURL, err := url.Parse(config.TargetURL)
	if err != nil {
		slog.Error("failed to parse target URL", "error", err)
		return
	}

	var ca cache.Cache

	if strings.ToLower(config.CacheType) == "disk" {
		ca = cache.NewDiskBackedCache(config.Ttl)
	} else {
		ca = cache.NewInAppCache(config.Ttl, config.MaxMemoryMB)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	handler := middleware.CompressMiddleware(config.Ttl, ca, zs, gz, proxy)

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(config.Port),
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("Server is running", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server crashed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited gracefully.")

}
