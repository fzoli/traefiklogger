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
	Enabled          bool      `json:"enabled"`
	LogFormat        LogFormat `json:"logFormat"`
	Name             string    `json:"name,omitempty"`
	BodyContentTypes []string  `json:"bodyContentTypes,omitempty"`
}

// LogFormat specifies the log format.
type LogFormat string

const (
	// TextFormat indicates text log format.
	TextFormat LogFormat = "text"
	// JSONFormat indicates JSON log format.
	JSONFormat LogFormat = "json"
)

// NoOpMiddleware a no-op plugin implementation.
type NoOpMiddleware struct {
	next http.Handler
}

func (m *NoOpMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.next.ServeHTTP(w, r)
}

// HTTPLogger a logger strategy interface.
type HTTPLogger interface {
	// print Prints the HTTP log.
	print(system string, r *http.Request, mrw *multiResponseWriter, requestHeaders string, requestBody *bytes.Buffer, responseHeaders string)
}

// LoggerMiddleware a Logger plugin.
type LoggerMiddleware struct {
	name         string
	logger       HTTPLogger
	contentTypes []string
	next         http.Handler
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Enabled:          true,
		LogFormat:        TextFormat,
		Name:             "HTTP",
		BodyContentTypes: []string{},
	}
}

// New creates a new LoggerMiddleware plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if !config.Enabled {
		return &NoOpMiddleware{
			next: next,
		}, nil
	}
	return &LoggerMiddleware{
		name:         config.Name,
		logger:       createHTTPLogger(ctx, config),
		contentTypes: config.BodyContentTypes,
		next:         next,
	}, nil
}

func createHTTPLogger(ctx context.Context, config *Config) HTTPLogger {
	logger := log.New(os.Stdout, "["+config.Name+"] ", log.LstdFlags)
	switch config.LogFormat {
	case JSONFormat:
		return createJsonHTTPLogger(ctx, logger)
	default:
		return createTextualHTTPLogger(ctx, logger)
	}
}

func (m *LoggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestWithBody := len(m.contentTypes) == 0
	for _, contentType := range m.contentTypes {
		if strings.Contains(r.Header.Get("Content-Type"), contentType) {
			requestWithBody = true
			break
		}
	}

	responseWithBody := len(m.contentTypes) == 0
	for _, contentType := range m.contentTypes {
		if strings.Contains(r.Header.Get("Accept"), contentType) {
			responseWithBody = true
			break
		}
	}

	requestBody := &bytes.Buffer{}
	mrc := &multiReadCloser{
		rc:       r.Body,
		buf:      requestBody,
		withBody: requestWithBody,
	}
	r.Body = mrc

	mrw := &multiResponseWriter{
		ResponseWriter: w,
		status:         200, // Default is 200
		body:           &bytes.Buffer{},
		withBody:       responseWithBody,
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

	m.logger.print(m.name, r, mrw, requestHeaders, requestBody, responseHeaders)
}

type multiResponseWriter struct {
	http.ResponseWriter
	status   int
	length   int
	body     *bytes.Buffer
	withBody bool
}

func (w *multiResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
}

func (w *multiResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	if w.withBody {
		w.body.Write(b)
	}
	return n, err
}

type multiReadCloser struct {
	rc       io.ReadCloser
	buf      *bytes.Buffer
	withBody bool
}

func (mrc *multiReadCloser) Read(p []byte) (int, error) {
	n, err := mrc.rc.Read(p)
	if mrc.withBody && n > 0 {
		mrc.buf.Write(p[:n])
	}
	return n, err
}

func (mrc *multiReadCloser) Close() error {
	return mrc.rc.Close()
}
