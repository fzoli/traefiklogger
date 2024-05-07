// Package traefiklogger a Traefik HTTP logger plugin.
package traefiklogger

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// JSONHTTPLogger a JSON logger implementation.
type JSONHTTPLogger struct {
	logger *log.Logger
}

func (jhl *JSONHTTPLogger) print(system string, r *http.Request, mrw *multiResponseWriter, requestHeaders string, requestBody *bytes.Buffer, responseHeaders string) {
	logData := struct {
		System             string            `json:"system,omitempty"`
		Time               string            `json:"time"`
		RemoteAddr         string            `json:"remoteAddr,omitempty"`
		Method             string            `json:"method"`
		URL                string            `json:"url"`
		Status             int               `json:"status"`
		StatusText         string            `json:"statusText"`
		Proto              string            `json:"proto"`
		RequestHeaders     map[string]string `json:"requestHeaders,omitempty"`
		RequestBody        string            `json:"requestBody,omitempty"`
		ResponseHeaders    map[string]string `json:"responseHeaders,omitempty"`
		ResponseContentLen int               `json:"responseContentLength"`
		ResponseBody       string            `json:"responseBody,omitempty"`
	}{
		System:             system,
		Time:               time.Now().UTC().Format("2006-01-02T15:04:05.999Z07:00"),
		RemoteAddr:         r.RemoteAddr,
		Method:             r.Method,
		URL:                r.URL.String(),
		Status:             mrw.status,
		StatusText:         http.StatusText(mrw.status),
		Proto:              r.Proto,
		RequestHeaders:     parseHeaders(requestHeaders),
		RequestBody:        requestBody.String(),
		ResponseHeaders:    parseHeaders(responseHeaders),
		ResponseContentLen: mrw.length,
		ResponseBody:       mrw.body.String(),
	}

	logBytes, err := json.Marshal(logData)
	if err != nil {
		jhl.logger.Println("Failed to marshal json log data")
		return
	}

	_, err = os.Stdout.WriteString(string(logBytes) + "\n")
	if err != nil {
		jhl.logger.Println("Failed to print")
		return
	}
}

func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	lines := strings.Split(headerStr, "\n")
	for _, line := range lines {
		if line != "" {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				headers[parts[0]] = parts[1]
			}
		}
	}
	return headers
}
