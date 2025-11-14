package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// MakeRequest creates an HTTP request for testing
func MakeRequest(method, url string, body interface{}) (*http.Request, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// ParseJSONResponse parses a JSON response body into the target struct
func ParseJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, target interface{}) {
	t.Helper()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("failed to decode response body: %v\nBody: %s", err, resp.Body.String())
	}
}

// AssertStatus checks that the response has the expected status code
func AssertStatus(t *testing.T, resp *httptest.ResponseRecorder, expected int) {
	t.Helper()

	if resp.Code != expected {
		t.Errorf("expected status %d, got %d\nBody: %s", expected, resp.Code, resp.Body.String())
	}
}

// AssertJSONEqual compares two JSON structures for equality
// It provides detailed diff output if they don't match
func AssertJSONEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("JSON mismatch (-expected +actual):\n%s", diff)
		if len(msgAndArgs) > 0 {
			t.Errorf("Additional context: %v", msgAndArgs...)
		}
	}
}

// AssertJSONContains checks if the actual JSON contains all fields from expected
// Useful for partial matching when you don't want to check all fields
func AssertJSONContains(t *testing.T, expected, actual map[string]interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	for key, expectedValue := range expected {
		actualValue, ok := actual[key]
		if !ok {
			t.Errorf("expected key %q not found in actual response", key)
			continue
		}

		if diff := cmp.Diff(expectedValue, actualValue); diff != "" {
			t.Errorf("value mismatch for key %q (-expected +actual):\n%s", key, diff)
		}
	}
}

// AssertHeader checks that the response has the expected header value
func AssertHeader(t *testing.T, resp *httptest.ResponseRecorder, header, expected string) {
	t.Helper()

	actual := resp.Header().Get(header)
	if actual != expected {
		t.Errorf("expected header %s=%q, got %q", header, expected, actual)
	}
}

// AssertContentType checks that the response has the expected Content-Type header
func AssertContentType(t *testing.T, resp *httptest.ResponseRecorder, expected string) {
	t.Helper()

	AssertHeader(t, resp, "Content-Type", expected)
}

// AssertErrorResponse checks that the response is an error with the expected code and message
func AssertErrorResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedCode string, expectedMessage string) {
	t.Helper()

	var errorResp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	ParseJSONResponse(t, resp, &errorResp)

	if errorResp.Error.Code != expectedCode {
		t.Errorf("expected error code %q, got %q", expectedCode, errorResp.Error.Code)
	}

	if expectedMessage != "" && errorResp.Error.Message != expectedMessage {
		t.Errorf("expected error message %q, got %q", expectedMessage, errorResp.Error.Message)
	}
}

