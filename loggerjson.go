// Package traefiklogger a Traefik HTTP logger plugin.
package traefiklogger

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// JSONHTTPLogger a JSON logger implementation.
type JSONHTTPLogger struct {
	clock   LoggerClock
	logger  *log.Logger
	writer  LogWriter
}

type Origin struct {
	FileName string `json:"file.name,omitempty"`
	FileLine int    `json:"file.line,omitempty"`
}

func (jhl *JSONHTTPLogger) print(record *LogRecord) {
	logData := struct {
		Level                 string              `json:"log.level,omitempty"`
		Time                  string              `json:"@timestamp"`
		Origin                Origin              `json:"log.origin"`
		Message               string              `json:"message,omitempty"`
		System                string              `json:"system,omitempty"`
		RemoteAddr            string              `json:"remoteAddr,omitempty"`
		Method                string              `json:"method"`
		URL                   string              `json:"url"`
		Status                int                 `json:"status"`
		StatusText            string              `json:"statusText"`
		Proto                 string              `json:"proto"`
		RequestHeaders        map[string][]string `json:"requestHeaders,omitempty"`
		RequestBody           string              `json:"requestBody,omitempty"`
		ResponseHeaders       map[string][]string `json:"responseHeaders,omitempty"`
		ResponseContentLength int                 `json:"responseContentLength"`
		ResponseBody          string              `json:"responseBody,omitempty"`
		EcsVersion            string              `json:"ecs.version,omitempty"`
	}{
		Level:                 "info",
		Time:                  jhl.clock.Now().UTC().Format("2006-01-02T15:04:05.999Z07:00"),
		Origin:                Origin{FileName: "loggerjson.go", FileLine: 1},
		Message:               fmt.Sprintf("%s %s %s %d", record.Method, record.URL, record.Proto, record.StatusCode),
		System:                record.System,
		RemoteAddr:            record.RemoteAddr,
		Method:                record.Method,
		URL:                   record.URL,
		Status:                record.StatusCode,
		StatusText:            http.StatusText(record.StatusCode),
		Proto:                 record.Proto,
		RequestHeaders:        record.RequestHeaders,
		RequestBody:           record.RequestBody.String(),
		ResponseHeaders:       record.ResponseHeaders,
		ResponseContentLength: record.ResponseContentLength,
		ResponseBody:          record.ResponseBody.String(),
		EcsVersion:            "1.6.0",
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
