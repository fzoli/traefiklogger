// Package traefiklogger a Traefik HTTP logger plugin.
package traefiklogger

import (
	"bytes"
	"context"
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
	print(system string, r *http.Request, mrw *multiResponseWriter, requestHeaders http.Header, requestBody *bytes.Buffer, responseHeaders http.Header)
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

	logger := log.New(os.Stdout, "["+config.Name+"] ", log.LstdFlags)
	var httpLogger HTTPLogger
	switch config.LogFormat {
	case JSONFormat:
		httpLogger = createJSONHTTPLogger(ctx, logger)
	default:
		httpLogger = createTextualHTTPLogger(ctx, logger)
	}

	return &LoggerMiddleware{
		name:         config.Name,
		logger:       httpLogger,
		contentTypes: config.BodyContentTypes,
		next:         next,
	}, nil
}

func (m *LoggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestBody := &bytes.Buffer{}
	mrc := &multiReadCloser{
		rc:       r.Body,
		buf:      requestBody,
		withBody: needToLogRequestBody(m, r),
	}
	r.Body = mrc

	mrw := &multiResponseWriter{
		ResponseWriter: w,
		status:         200, // Default is 200
		body:           &bytes.Buffer{},
		withBody:       needToLogResponseBody(m, r),
	}

	requestHeaders := copyHeaders(r.Header)

	m.next.ServeHTTP(mrw, r)

	responseHeaders := copyHeaders(w.Header())

	m.logger.print(m.name, r, mrw, requestHeaders, requestBody, responseHeaders)
}

func needToLogRequestBody(m *LoggerMiddleware, r *http.Request) bool {
	for _, contentType := range m.contentTypes {
		if strings.Contains(r.Header.Get("Content-Type"), contentType) {
			return true
		}
	}
	return len(m.contentTypes) == 0
}

func needToLogResponseBody(m *LoggerMiddleware, r *http.Request) bool {
	for _, contentType := range m.contentTypes {
		if strings.Contains(r.Header.Get("Accept"), contentType) {
			return true
		}
	}
	return len(m.contentTypes) == 0
}

func copyHeaders(original http.Header) http.Header {
	newHeader := make(http.Header)
	for key, value := range original {
		newHeader[key] = value
	}
	return newHeader
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
