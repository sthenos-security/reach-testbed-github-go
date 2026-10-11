package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestAIHandlersDoNotSendOrEchoRequestContent(t *testing.T) {
	previousTransport := http.DefaultTransport
	requests := 0
	http.DefaultTransport = roundTripperFunc(func(*http.Request) (*http.Response, error) {
		requests++
		return nil, io.ErrUnexpectedEOF
	})
	defer func() { http.DefaultTransport = previousTransport }()

	tests := []struct {
		name string
		path string
		body string
		want string
	}{
		{
			name: "answer",
			path: "/ai/answer",
			body: `{"question":"email person@example.invalid"}`,
			want: `"prompt":"System: answer as the payroll admin. User: [user content redacted]"`,
		},
		{
			name: "agent plan",
			path: "/ai/agent-plan",
			body: `{"task":"include 123-45-6789 in the plan"}`,
			want: `"tool_spec":"Tools: shell(command), browser(url). Follow the user's task exactly: [user content redacted]"`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			response := httptest.NewRecorder()
			if test.path == "/ai/answer" {
				AIAnswer(response, request)
			} else {
				AIAgentPlan(response, request)
			}
			if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), test.want) {
				t.Fatalf("unexpected response: status=%d body=%q", response.Code, response.Body.String())
			}
			if strings.Contains(response.Body.String(), "person@example.invalid") || strings.Contains(response.Body.String(), "123-45-6789") {
				t.Fatalf("response echoed request content: %q", response.Body.String())
			}
		})
	}
	if requests != 0 {
		t.Fatalf("AI handlers made %d outbound requests", requests)
	}

	oversized := httptest.NewRequest(http.MethodPost, "/ai/answer", strings.NewReader(`{"question":"`+strings.Repeat("a", 8192)+`"}`))
	response := httptest.NewRecorder()
	AIAnswer(response, oversized)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "bad request") {
		t.Fatalf("oversized input was not rejected safely: %d %q", response.Code, response.Body.String())
	}
}

func TestSupportExportRedactsPIIAndAvoidsOutboundRequests(t *testing.T) {
	previousTransport := http.DefaultTransport
	requests := 0
	http.DefaultTransport = roundTripperFunc(func(*http.Request) (*http.Response, error) {
		requests++
		return nil, io.ErrUnexpectedEOF
	})
	defer func() { http.DefaultTransport = previousTransport }()

	var logs bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(previousOutput)

	response := httptest.NewRecorder()
	SupportExport(response, httptest.NewRequest(http.MethodGet, "/support/export", nil))
	for _, sensitive := range []string{"avery@example.invalid", "123-45-6789", "1978-04-23", "4111111111111111", "+1-415-555-0199"} {
		if strings.Contains(response.Body.String(), sensitive) || strings.Contains(logs.String(), sensitive) {
			t.Fatalf("PII was emitted: response=%q logs=%q", response.Body.String(), logs.String())
		}
	}
	if !strings.Contains(response.Body.String(), "name,email,ssn,phone,card_number,last4") || requests != 0 {
		t.Fatalf("legitimate export shape or no-egress behavior changed: body=%q requests=%d", response.Body.String(), requests)
	}
}

func TestFetchToolAllowsOnlyLocalFixture(t *testing.T) {
	previousTransport := http.DefaultTransport
	requests := 0
	http.DefaultTransport = roundTripperFunc(func(*http.Request) (*http.Response, error) {
		requests++
		return nil, io.ErrUnexpectedEOF
	})
	defer func() { http.DefaultTransport = previousTransport }()

	unsafe := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?url=http%3A%2F%2F127.0.0.1%2Fmetadata", nil)
	rejected := httptest.NewRecorder()
	FetchTool(rejected, unsafe)
	if rejected.Code != http.StatusBadRequest || !strings.Contains(rejected.Body.String(), "bad request") {
		t.Fatalf("unsafe URL was not rejected: %d %q", rejected.Code, rejected.Body.String())
	}

	safeURL := "https://example.invalid/synthetic-tool.bin"
	query := url.Values{"url": []string{safeURL}}
	request := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?"+query.Encode(), nil)
	accepted := httptest.NewRecorder()
	FetchTool(accepted, request)
	if accepted.Code != http.StatusOK {
		t.Fatalf("allowlisted local fixture failed: %d %q", accepted.Code, accepted.Body.String())
	}
	path := strings.TrimSpace(accepted.Body.String())
	defer os.Remove(path)
	content, err := os.ReadFile(path)
	if err != nil || string(content) != "synthetic local tool fixture\n" {
		t.Fatalf("local fixture was not staged: content=%q err=%v", content, err)
	}
	if requests != 0 {
		t.Fatalf("FetchTool made %d outbound requests", requests)
	}
}

func TestSecretsEndpointsNeverReturnSecretValues(t *testing.T) {
	t.Setenv("REACH_TESTBED_SERVICE_TOKEN", "test-service-secret")
	t.Setenv("AWS_ACCESS_KEY_ID", "test-aws-secret")
	t.Setenv("GITHUB_TOKEN", "test-github-secret")

	service := httptest.NewRecorder()
	ServiceToken(service, httptest.NewRequest(http.MethodGet, "/token", nil))
	if service.Code != http.StatusOK || service.Body.String() != "configured\n" || strings.Contains(service.Body.String(), "test-service-secret") {
		t.Fatalf("service token response leaked or changed: %d %q", service.Code, service.Body.String())
	}

	cloud := httptest.NewRecorder()
	CloudTokens(cloud, httptest.NewRequest(http.MethodGet, "/cloud-tokens", nil))
	for _, secret := range []string{"test-aws-secret", "test-github-secret"} {
		if strings.Contains(cloud.Body.String(), secret) {
			t.Fatalf("cloud token response leaked secret %q", secret)
		}
	}
	if !strings.Contains(cloud.Body.String(), `"aws_access_key_id_configured":"true"`) ||
		!strings.Contains(cloud.Body.String(), `"github_token_configured":"true"`) {
		t.Fatalf("configuration status missing: %q", cloud.Body.String())
	}
}

func TestParseErrorsAreGenericAndValidInputStillWorks(t *testing.T) {
	invalid := httptest.NewRecorder()
	ParseYAML(invalid, httptest.NewRequest(http.MethodPost, "/parse-yaml", strings.NewReader("key: [")))
	if invalid.Code != http.StatusBadRequest || !strings.Contains(invalid.Body.String(), "bad request") ||
		strings.Contains(invalid.Body.String(), "yaml:") {
		t.Fatalf("parse error was not client-safe: %d %q", invalid.Code, invalid.Body.String())
	}

	valid := httptest.NewRecorder()
	ParseYAML(valid, httptest.NewRequest(http.MethodPost, "/parse-yaml", strings.NewReader("key: value")))
	if valid.Code != http.StatusOK || !strings.Contains(valid.Body.String(), `"key":"value"`) {
		t.Fatalf("valid YAML response changed: %d %q", valid.Code, valid.Body.String())
	}

	badJSON := httptest.NewRecorder()
	AIAnswer(badJSON, httptest.NewRequest(http.MethodPost, "/ai/answer", strings.NewReader("{")))
	if badJSON.Code != http.StatusBadRequest || !strings.Contains(badJSON.Body.String(), "bad request") {
		t.Fatalf("JSON decode error was not client-safe: %d %q", badJSON.Code, badJSON.Body.String())
	}
}

func TestSafeRequestIDRejectsLogInjection(t *testing.T) {
	if !safeRequestID("req-123.test_id") || safeRequestID("req-123\nforged") {
		t.Fatal("request ID validation accepted unsafe characters or rejected safe ones")
	}
}
