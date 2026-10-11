package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		statusCode int
		body       string
	}{
		{
			name:       "reject crafted tag",
			tag:        strings.Repeat("en-", 1000),
			statusCode: http.StatusBadRequest,
			body:       "unsupported language tag\n",
		},
		{
			name:       "parse supported tag",
			tag:        "en-US",
			statusCode: http.StatusOK,
			body:       "en-US\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := url.Values{"tag": []string{tt.tag}}.Encode()
			request := httptest.NewRequest(http.MethodPost, "/parse-language?"+query, nil)
			response := httptest.NewRecorder()

			ParseLanguage(response, request)

			if response.Code != tt.statusCode {
				t.Errorf("status code = %d, want %d", response.Code, tt.statusCode)
			}
			if response.Body.String() != tt.body {
				t.Errorf("body = %q, want %q", response.Body.String(), tt.body)
			}
		})
	}
}
