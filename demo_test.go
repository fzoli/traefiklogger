package plugindemo_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/traefik/plugindemo"
)

func TestPost(t *testing.T) {
	cfg := &plugindemo.Config{}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Read the request body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		// Parse the request body as an integer
		num, err := strconv.Atoi(string(body))
		if err != nil {
			http.Error(rw, "Bad Request: Request body must be an integer", http.StatusBadRequest)
			return
		}

		// Double the number
		result := num * 2

		// Write the result as the response body
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "%d", result)
	})

	handler, err := plugindemo.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/post", strings.NewReader("5"))
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	// Check the response body
	if recorder.Body.String() != "10" {
		t.Errorf("Expected response body: '10', got: '%s'", recorder.Body.String())
	}
}

func TestEmptyPost(t *testing.T) {
	cfg := &plugindemo.Config{}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Read the request body (to appear in logs)
		_, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := plugindemo.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/empty-post", strings.NewReader("5"))
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	// Check the response body
	if recorder.Body.String() != "" {
		t.Errorf("Expected response body: '', got: '%s'", recorder.Body.String())
	}
}

func TestGet(t *testing.T) {
	cfg := &plugindemo.Config{}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Write the result as the response body
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "%d", 5)
	})

	handler, err := plugindemo.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/get", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	// Check the response body
	if recorder.Body.String() != "5" {
		t.Errorf("Expected response body: '5', got: '%s'", recorder.Body.String())
	}
}

func TestEmptyGet(t *testing.T) {
	cfg := &plugindemo.Config{}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := plugindemo.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/empty-get", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	// Check the response body
	if recorder.Body.String() != "" {
		t.Errorf("Expected response body: '', got: '%s'", recorder.Body.String())
	}
}
