package handlers

import (
	"net/http/httptest"
	"net/url"
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
			name:       "unsupported tag is rejected",
			tag:        "de",
			statusCode: 400,
			body:       "unsupported language tag\n",
		},
		{
			name:       "supported tag is parsed",
			tag:        "en-US",
			statusCode: 200,
			body:       "en-US\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := url.Values{"tag": []string{test.tag}}.Encode()
			request := httptest.NewRequest("POST", "/parse-language?"+query, nil)
			response := httptest.NewRecorder()

			ParseLanguage(response, request)

			if response.Code != test.statusCode {
				t.Errorf("status code = %d, want %d", response.Code, test.statusCode)
			}
			if response.Body.String() != test.body {
				t.Errorf("body = %q, want %q", response.Body.String(), test.body)
			}
		})
	}
}
