package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/sirupsen/logrus"
)

// testLogger creates a logger for testing that discards output
func testLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logger
}

// Mock types for testing

// Mock types are defined in validation_test.go and reused here

// Helper functions for testing

func createTestRequest(method, path string, body interface{}) *http.Request {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBody)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func executeTestRequest(server *Server, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	server.getRouter().ServeHTTP(w, req)
	return w
}

func decodeJSON(body io.Reader, v interface{}) error {
	return json.NewDecoder(body).Decode(v)
}
