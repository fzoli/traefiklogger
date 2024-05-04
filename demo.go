// Package plugindemo a demo plugin.
package plugindemo

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"
)

// Config the plugin configuration.
type Config struct {
}

// LoggerMiddleware a Logger plugin.
type LoggerMiddleware struct {
	logger *log.Logger
	next   http.Handler
}

// New creates a new LoggerMiddleware plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
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
	m.next.ServeHTTP(mrw, r)

	m.logger.Printf("%s %s %s : Status %d %s\nRequest Body: %s\nResponse Body: %s\nResponse Content Length: %d\n",
		r.RemoteAddr, r.Method, r.URL.String(),
		mrw.status, http.StatusText(mrw.status),
		requestBody.String(), mrw.body.String(),
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
