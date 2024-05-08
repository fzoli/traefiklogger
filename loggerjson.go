// Package traefiklogger a Traefik HTTP logger plugin.
package traefiklogger

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

// JSONHTTPLogger a JSON logger implementation.
type JSONHTTPLogger struct {
	clock  LoggerClock
	logger *log.Logger
	writer LogWriter
}

func (jhl *JSONHTTPLogger) print(system string, r *http.Request, mrw *multiResponseWriter, requestHeaders http.Header, requestBody *bytes.Buffer, responseHeaders http.Header) {
	logData := struct {
		System             string              `json:"system,omitempty"`
		Time               string              `json:"time"`
		RemoteAddr         string              `json:"remoteAddr,omitempty"`
		Method             string              `json:"method"`
		URL                string              `json:"url"`
		Status             int                 `json:"status"`
		StatusText         string              `json:"statusText"`
		Proto              string              `json:"proto"`
		RequestHeaders     map[string][]string `json:"requestHeaders,omitempty"`
		RequestBody        string              `json:"requestBody,omitempty"`
		ResponseHeaders    map[string][]string `json:"responseHeaders,omitempty"`
		ResponseContentLen int                 `json:"responseContentLength"`
		ResponseBody       string              `json:"responseBody,omitempty"`
	}{
		System:             system,
		Time:               jhl.clock.Now().UTC().Format("2006-01-02T15:04:05.999Z07:00"),
		RemoteAddr:         r.RemoteAddr,
		Method:             r.Method,
		URL:                r.URL.String(),
		Status:             mrw.status,
		StatusText:         http.StatusText(mrw.status),
		Proto:              r.Proto,
		RequestHeaders:     requestHeaders,
		RequestBody:        requestBody.String(),
		ResponseHeaders:    responseHeaders,
		ResponseContentLen: mrw.length,
		ResponseBody:       mrw.body.String(),
	}

	logBytes, err := json.Marshal(logData)
	if err != nil {
		jhl.logger.Println("Failed to marshal json log data")
		return
	}

	err = jhl.writer.Write(string(logBytes) + "\n")
	if err != nil {
		jhl.logger.Println("Failed to write:", err)
		return
	}
}
