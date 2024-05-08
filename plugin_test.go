package traefiklogger_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fzoli/traefiklogger"
)

type TestLoggerClock struct{}

func (c *TestLoggerClock) Now() time.Time {
	return time.Date(2020, time.December, 15, 13, 30, 40, 999999999, time.UTC)
}

type TestLogWriter struct {
	t        *testing.T
	expected string
}

func (w *TestLogWriter) Write(log string) error {
	w.t.Helper()
	if log != w.expected {
		w.t.Errorf("Expected: '%s', got: '%s'", w.expected, log)
	}
	return nil
}

// createContext creates text context with fake time and test log writer that assert the expected log.
func createContext(t *testing.T, expectedLog string) context.Context {
	t.Helper()
	return context.WithValue(context.WithValue(context.Background(), traefiklogger.LogWriterContextKey, &TestLogWriter{t: t, expected: expectedLog}), traefiklogger.ClockContextKey, &TestLoggerClock{})
}

// doubleTheNumber reads the request, parses it as integer then returns its double.
// So the request and the response are not the same.
func doubleTheNumber(rw http.ResponseWriter, req *http.Request) {
	// Read the request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		cerr := Body.Close()
		if cerr != nil {
			log.Printf("Failed to close reader: %v", cerr)
		}
	}(req.Body)

	// Parse the request body as an integer
	num, err := strconv.Atoi(string(body))
	if err != nil {
		http.Error(rw, "Bad Request: Request body must be an integer", http.StatusBadRequest)
		return
	}

	// Double the number
	result := num * 2

	// Write the result
	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(rw, "%d", result)
}

// blackHole reads the request then it just returns HTTP OK without response body.
func blackHole(rw http.ResponseWriter, req *http.Request) {
	// Read the request body (to appear in logs)
	_, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		cerr := Body.Close()
		if cerr != nil {
			log.Printf("Failed to close reader: %v", cerr)
		}
	}(req.Body)
	rw.WriteHeader(http.StatusOK)
}

// alwaysFive does not read the request, just returns HTTP OK with response body 5.
func alwaysFive(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, "%d", 5)
}

func TestPost(t *testing.T) {
	expectedLogs := map[traefiklogger.LogFormat]string{
		traefiklogger.TextFormat: "127.0.0.1 POST http://localhost/post: 200 OK HTTP/1.1\n\nRequest Headers:\nAccept: text/plain\n\nRequest Body:\n5\n\nResponse Headers:\nContent-Type: text/plain\n\nResponse Content Length: 2\n\nResponse Body:\n10\n\n",
		traefiklogger.JSONFormat: "{\"log.level\":\"info\",\"@timestamp\":\"2020-12-15T13:30:40.999Z\",\"message\":\"POST http://localhost/post HTTP/1.1 200\",\"systemName\":\"HTTP\",\"remoteAddress\":\"127.0.0.1\",\"method\":\"POST\",\"path\":\"http://localhost/post\",\"status\":200,\"statusText\":\"OK\",\"proto\":\"HTTP/1.1\",\"requestHeaders\":{\"Accept\":[\"text/plain\"]},\"requestBody\":\"5\",\"responseHeaders\":{\"Content-Type\":[\"text/plain\"]},\"responseContentLength\":2,\"responseBody\":\"10\",\"ecs.version\":\"1.6.0\"}\n",
	}

	for logFormat, expectedLog := range expectedLogs {
		cfg := traefiklogger.CreateConfig()
		cfg.LogFormat = logFormat

		ctx := createContext(t, expectedLog)

		handler, err := traefiklogger.New(ctx, http.HandlerFunc(doubleTheNumber), cfg, "logger-plugin")
		if err != nil {
			t.Fatal(err)
		}

		recorder := httptest.NewRecorder()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/post", strings.NewReader("5"))
		if err != nil {
			t.Fatal(err)
		}
		req.RemoteAddr = "127.0.0.1"
		req.Header.Set("Accept", "text/plain")

		handler.ServeHTTP(recorder, req)

		// Check the response body
		if recorder.Body.String() != "10" {
			t.Errorf("Expected response body: '10', got: '%s'", recorder.Body.String())
		}
	}
}

func TestShortPost(t *testing.T) {
	expectedLogs := map[traefiklogger.LogFormat]string{
		traefiklogger.TextFormat: "127.0.0.1 POST http://localhost/short-post: 200 OK HTTP/1.1\n\nRequest Headers:\nAccept: text/plain\n\nResponse Headers:\nContent-Type: text/plain\n\nResponse Content Length: 2\n\n",
		traefiklogger.JSONFormat: "{\"log.level\":\"info\",\"@timestamp\":\"2020-12-15T13:30:40.999Z\",\"message\":\"POST http://localhost/short-post HTTP/1.1 200\",\"systemName\":\"HTTP\",\"remoteAddress\":\"127.0.0.1\",\"method\":\"POST\",\"path\":\"http://localhost/short-post\",\"status\":200,\"statusText\":\"OK\",\"proto\":\"HTTP/1.1\",\"requestHeaders\":{\"Accept\":[\"text/plain\"]},\"responseHeaders\":{\"Content-Type\":[\"text/plain\"]},\"responseContentLength\":2,\"ecs.version\":\"1.6.0\"}\n",
	}

	for logFormat, expectedLog := range expectedLogs {
		cfg := traefiklogger.CreateConfig()
		cfg.LogFormat = logFormat
		cfg.BodyContentTypes = []string{"text/html"}

		ctx := createContext(t, expectedLog)

		handler, err := traefiklogger.New(ctx, http.HandlerFunc(doubleTheNumber), cfg, "logger-plugin")
		if err != nil {
			t.Fatal(err)
		}

		recorder := httptest.NewRecorder()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/short-post", strings.NewReader("5"))
		if err != nil {
			t.Fatal(err)
		}
		req.RemoteAddr = "127.0.0.1"
		req.Header.Set("Accept", "text/plain")

		handler.ServeHTTP(recorder, req)

		// Check the response body
		if recorder.Body.String() != "10" {
			t.Errorf("Expected response body: '10', got: '%s'", recorder.Body.String())
		}
	}
}

func TestEmptyPost(t *testing.T) {
	cfg := traefiklogger.CreateConfig()

	ctx := context.WithValue(context.Background(), traefiklogger.LogWriterContextKey, &TestLogWriter{t: t, expected: "127.0.0.1 POST http://localhost/empty-post: 200 OK HTTP/1.1\n\nRequest Body:\n5\n\nResponse Content Length: 0\n\n"})

	handler, err := traefiklogger.New(ctx, http.HandlerFunc(blackHole), cfg, "logger-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/empty-post", strings.NewReader("5"))
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1"

	handler.ServeHTTP(recorder, req)

	// Check the response body
	if recorder.Body.String() != "" {
		t.Errorf("Expected response body: '', got: '%s'", recorder.Body.String())
	}
}

func TestGet(t *testing.T) {
	cfg := traefiklogger.CreateConfig()

	ctx := context.WithValue(context.Background(), traefiklogger.LogWriterContextKey, &TestLogWriter{t: t, expected: "127.0.0.1 GET http://localhost/get: 200 OK HTTP/1.1\n\nResponse Content Length: 1\n\nResponse Body:\n5\n\n"})

	handler, err := traefiklogger.New(ctx, http.HandlerFunc(alwaysFive), cfg, "logger-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/get", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1"

	handler.ServeHTTP(recorder, req)

	// Check the response body
	if recorder.Body.String() != "5" {
		t.Errorf("Expected response body: '5', got: '%s'", recorder.Body.String())
	}
}

func TestEmptyGet(t *testing.T) {
	cfg := traefiklogger.CreateConfig()

	ctx := context.WithValue(context.Background(), traefiklogger.LogWriterContextKey, &TestLogWriter{t: t, expected: "127.0.0.1 GET http://localhost/empty-get: 200 OK HTTP/1.1\n\nResponse Content Length: 0\n\n"})
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := traefiklogger.New(ctx, next, cfg, "logger-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/empty-get", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1"

	handler.ServeHTTP(recorder, req)

	// Check the response body
	if recorder.Body.String() != "" {
		t.Errorf("Expected response body: '', got: '%s'", recorder.Body.String())
	}
}

func TestDisabled(t *testing.T) {
	cfg := traefiklogger.CreateConfig()
	cfg.Enabled = false

	ctx := context.WithValue(context.Background(), traefiklogger.LogWriterContextKey, &TestLogWriter{t: t, expected: ""})

	handler, err := traefiklogger.New(ctx, http.HandlerFunc(alwaysFive), cfg, "logger-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/disabled", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1"

	handler.ServeHTTP(recorder, req)

	// Check the response body
	if recorder.Body.String() != "5" {
		t.Errorf("Expected response body: '5', got: '%s'", recorder.Body.String())
	}
}
