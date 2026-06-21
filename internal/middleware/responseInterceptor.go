package middleware

import (
	"bytes"
	"net/http"
)

type ResponseInterceptor struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
}

func (ri *ResponseInterceptor) WriteHeader(statusCode int) {
	ri.StatusCode = statusCode
}

func (ri *ResponseInterceptor) Write(b []byte) (int, error) {
	return ri.Body.Write(b)
}