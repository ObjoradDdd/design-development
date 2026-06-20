package main

import (
	"design-developer/internal/config"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {

	wg := sync.WaitGroup{}

	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	targetURL, err := url.Parse(config.TargetURL)
	if err != nil {
		slog.Error("failed to parse target URL", "error", err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(config.Port),
		Handler:      proxy,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("Server is running", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server crashed", "error", err)
			os.Exit(1)
		}
	}()


	wg.Wait()
}