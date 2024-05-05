// Package traefiklogger a Traefik HTTP logger plugin.
package traefiklogger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Config the plugin configuration.
type Config struct {
	Enabled bool `json:"enabled"`
}

// NoOpMiddleware a no-op plugin implementation.
type NoOpMiddleware struct {
	next http.Handler
}

func (m *NoOpMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.next.ServeHTTP(w, r)
}

// LoggerMiddleware a Logger plugin.
type LoggerMiddleware struct {
	logger *log.Logger
	next   http.Handler
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Enabled: true,
	}
}

// New creates a new LoggerMiddleware plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if !config.Enabled {
		return &NoOpMiddleware{
			next: next,
		}, nil
	}
	logger := log.New(os.Stdout, "[HTTP] ", log.LstdFlags)
	return &LoggerMiddleware{
		logger: logger,
		next:   next,
	}, nil
}

func (m *LoggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestBody := &bytes.Buffer{}

	mrc := &multiReadCloser{
		rc:  r.Body,
		buf: requestBody,
	}
	r.Body = mrc

	mrw := &multiResponseWriter{
		ResponseWriter: w,
		status:         200, // Default is 200
		body:           &bytes.Buffer{},
	}

	requestHeaders := ""
	for key, values := range r.Header {
		requestHeaders += fmt.Sprintf("%s: %s\n", key, strings.Join(values, ","))
	}

	m.next.ServeHTTP(mrw, r)

	responseHeaders := ""
	for key, values := range w.Header() {
		responseHeaders += fmt.Sprintf("%s: %s\n", key, strings.Join(values, ","))
	}

	m.logger.Printf("%s %s %s: %d %s %s\n\nRequest Headers:\n%s\nRequest Body:\n%s\n\nResponse Headers:\n%s\nResponse Body:\n%s\n\nResponse Content Length: %d\n\n",
		r.RemoteAddr, r.Method, r.URL.String(),
		mrw.status, http.StatusText(mrw.status), r.Proto,
		requestHeaders, requestBody.String(),
		responseHeaders, mrw.body.String(),
		mrw.length,
	)
}

type multiResponseWriter struct {
	http.ResponseWriter
	status int
	length int
	body   *bytes.Buffer
}

func (w *multiResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
}

func (w *multiResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	w.body.Write(b)
	return n, err
}

type multiReadCloser struct {
	rc  io.ReadCloser
	buf *bytes.Buffer
}

func (mrc *multiReadCloser) Read(p []byte) (int, error) {
	n, err := mrc.rc.Read(p)
	if n > 0 {
		mrc.buf.Write(p[:n])
	}
	return n, err
}

func (mrc *multiReadCloser) Close() error {
	return mrc.rc.Close()
}
