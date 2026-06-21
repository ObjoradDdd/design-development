package middleware

import (
	"bytes"
	"design-developer/internal/cache"
	"design-developer/internal/compressor"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	GzipCompressorName = "gzip"
	ZstdCompressorName = "zstd"
)

func CompressMiddleware(defaultTtl time.Duration, ca cache.Cache, zs *compressor.ZstdCompressor, gz *compressor.GzipCompressor, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		acceptEncoding := r.Header.Get("Accept-Encoding")
		var encoding string

		if strings.Contains(acceptEncoding, ZstdCompressorName) && zs != nil {
			encoding = ZstdCompressorName
		} else if strings.Contains(acceptEncoding, GzipCompressorName) && gz != nil {
			encoding = GzipCompressorName
		}

		isCacheableMethod := r.Method == http.MethodGet || r.Method == http.MethodHead
		key := keyGen(r, encoding)

		if isCacheableMethod {
			if cachedData, found := ca.Get(key); found {

				for key, values := range cachedData.Headers {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}

				if encoding != "" {
					w.Header().Set("Content-Encoding", encoding)
				}
				w.Header().Set("Content-Length", strconv.Itoa(len(cachedData.Body)))
				w.WriteHeader(cachedData.StatusCode)
				w.Write(cachedData.Body)
				return
			}
		}

		interceptor := &ResponseInterceptor{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
			Body:           &bytes.Buffer{},
		}

		next.ServeHTTP(interceptor, r)

		rawData := interceptor.Body.Bytes()
		var finalData []byte

		if interceptor.Header().Get("Content-Encoding") == "" {
			if encoding == ZstdCompressorName {
				if compressed, err := zs.Compress(rawData); err == nil {
					finalData = compressed
				}
			} else if encoding == GzipCompressorName {
				if compressed, err := gz.Compress(rawData); err == nil {
					finalData = compressed
				}
			}
		}

		if finalData != nil {
			slog.Debug("compressed response for key", "key", key, "original_size", len(rawData), "compressed_size", len(finalData), "encoding", encoding)
			w.Header().Set("Content-Encoding", encoding)
		} else {
			finalData = rawData
		}

		if isCacheableMethod {
			ttl := parseTTL(interceptor.Header().Get("Cache-Control"), defaultTtl)

			if ttl > 0 {
				ca.Set(key, cache.CacheEntry{
					StatusCode: interceptor.StatusCode,
					Headers:    interceptor.Header().Clone(),
					Body:       finalData,
				}, ttl)
			}
		}

		w.Header().Set("Content-Length", strconv.Itoa(len(finalData)))
		w.WriteHeader(interceptor.StatusCode)
		w.Write(finalData)
	})
}

func parseTTL(cacheControl string, defaultTTL time.Duration) time.Duration {
	if cacheControl == "" {
		return defaultTTL
	}

	if strings.Contains(cacheControl, "no-cache") || strings.Contains(cacheControl, "no-store") {
		return 0
	}

	parts := strings.Split(cacheControl, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "max-age=") {
			valStr := strings.TrimPrefix(part, "max-age=")
			if seconds, err := strconv.Atoi(valStr); err == nil {
				return time.Duration(seconds) * time.Second
			}
		}
	}

	return defaultTTL
}