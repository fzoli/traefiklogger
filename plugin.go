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
	Enabled            bool      `json:"enabled"`
	LogFormat          LogFormat `json:"logFormat"`
	GenerateLogID      bool      `json:"generateLogId,omitempty"`
	Name               string    `json:"name,omitempty"`
	BodyContentTypes   []string  `json:"bodyContentTypes,omitempty"`
	JWTHeaders         []string  `json:"jwtHeaders,omitempty"`
	HeaderRedacts      []string  `json:"headerRedacts,omitempty"`
	RequestBodyRedact  string    `json:"requestBodyRedact,omitempty"`
	ResponseBodyRedact string    `json:"responseBodyRedact,omitempty"`
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
	print(record *LogRecord)
}

// LogRecord contains the loggable data.
type LogRecord struct {
	System                string
	Proto                 string
	Method                string
	URL                   string
	RemoteAddr            string
	StatusCode            int
	RequestHeaders        http.Header
	RequestBody           *bytes.Buffer
	ResponseHeaders       http.Header
	ResponseBody          *bytes.Buffer
	ResponseContentLength int
	DurationMs            float64
	BodyDecoder           HTTPBodyDecoder
}

// LoggerMiddleware a Logger plugin.
type LoggerMiddleware struct {
	name                string
	clock               LoggerClock
	logger              HTTPLogger
	bodyDecoderFactory  *HTTPBodyDecoderFactory
	contentTypes        []string
	jwtHeaders          []string
	headerRedacts       []string
	requestBodyRedacts  []string
	responseBodyRedacts []string
	next                http.Handler
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Enabled:            true,
		LogFormat:          TextFormat,
		GenerateLogID:      true,
		Name:               "HTTP",
		BodyContentTypes:   []string{},
		JWTHeaders:         []string{},
		HeaderRedacts:      []string{},
		RequestBodyRedact:  "",
		ResponseBodyRedact: "",
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

	return &LoggerMiddleware{
		name:                config.Name,
		clock:               createClock(ctx),
		logger:              createHTTPLogger(ctx, config, logger),
		bodyDecoderFactory:  createHTTPBodyDecoderFactory(logger),
		contentTypes:        config.BodyContentTypes,
		jwtHeaders:          config.JWTHeaders,
		headerRedacts:       config.HeaderRedacts,
		requestBodyRedacts:  strings.Split(config.RequestBodyRedact, ";"),
		responseBodyRedacts: strings.Split(config.ResponseBodyRedact, ";"),
		next:                next,
	}, nil
}

func (m *LoggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") == "websocket" {
		m.next.ServeHTTP(w, r)
		return
	}

	mrc := &multiReadCloser{
		rc:       r.Body,
		buf:      &bytes.Buffer{},
		withBody: !hasRedactedBody(r, m.requestBodyRedacts) && needToLogBody(m, r, "Content-Type"),
	}
	r.Body = mrc

	mrw := &multiResponseWriter{
		ResponseWriter: w,
		status:         200, // Default is 200
		body:           &bytes.Buffer{},
		withBody:       !hasRedactedBody(r, m.responseBodyRedacts) && needToLogBody(m, r, "Accept"),
	}

	requestHeaders := m.copyHeaders(r.Header)

	startTime := m.clock.Now()
	m.next.ServeHTTP(mrw, r)
	endTime := m.clock.Now()

	responseHeaders := m.copyHeaders(w.Header())
	durationMs := float64(endTime.UnixMicro()-startTime.UnixMicro()) / 1000.0
	bodyDecoder := m.bodyDecoderFactory.create(w.Header().Get("Content-Encoding"))

	logRecord := &LogRecord{
		System:                m.name,
		Proto:                 r.Proto,
		Method:                r.Method,
		URL:                   r.URL.String(),
		RemoteAddr:            r.RemoteAddr,
		StatusCode:            mrw.status,
		RequestHeaders:        requestHeaders,
		RequestBody:           mrc.buf,
		ResponseHeaders:       responseHeaders,
		ResponseBody:          mrw.body,
		ResponseContentLength: mrw.length,
		DurationMs:            durationMs,
		BodyDecoder:           bodyDecoder,
	}

	m.logger.print(logRecord)
}

func needToLogBody(m *LoggerMiddleware, r *http.Request, header string) bool {
	for _, contentType := range m.contentTypes {
		if strings.Contains(strings.ToLower(r.Header.Get(header)), strings.ToLower(contentType)) {
			return true
		}
	}
	return len(m.contentTypes) == 0
}

func hasRedactedBody(r *http.Request, redacts []string) bool {
	for _, requestBodyRedact := range redacts {
		if len(requestBodyRedact) == 0 {
			continue
		}
		method := r.Method + " " + r.URL.String()
		if strings.HasPrefix(method, requestBodyRedact) {
			return true
		}
	}
	return false
}

func (m *LoggerMiddleware) copyHeaders(original http.Header) http.Header {
	newHeader := make(http.Header)
	for key, value := range original {
		if containsIgnoreCase(m.headerRedacts, key) {
			newHeader[key] = decodeHeaders(value, redact)
			continue
		}
		if containsIgnoreCase(m.jwtHeaders, key) {
			newHeader[key] = decodeHeaders(value, decodeJWTHeader)
			continue
		}
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
