package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type requestRecorder struct {
	requests int
}

func (recorder *requestRecorder) RoundTrip(*http.Request) (*http.Response, error) {
	recorder.requests++
	return nil, errors.New("unexpected outbound HTTP request")
}

func handlerRequest(handler http.HandlerFunc, method, target, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	response := httptest.NewRecorder()
	handler(response, request)
	return response
}

func TestAIHandlersKeepRequestContentLocal(t *testing.T) {
	const privateInput = "PRIVATE-CONTENT-DO-NOT-LEAK-83917"
	transport := &requestRecorder{}
	originalTransport := http.DefaultTransport
	http.DefaultTransport = transport
	defer func() { http.DefaultTransport = originalTransport }()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		body    string
	}{
		{
			name:    "answer",
			handler: AIAnswer,
			body:    `{"question":"` + privateInput + `"}`,
		},
		{
			name:    "agent plan",
			handler: AIAgentPlan,
			body:    `{"task":"` + privateInput + `"}`,
		},
		{
			name:    "safe answer",
			handler: SafeAIAnswer,
			body:    `{"question":"A legitimate support question"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := handlerRequest(test.handler, http.MethodPost, "/", test.body)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if strings.Contains(response.Body.String(), privateInput) {
				t.Fatalf("response exposed request content: %s", response.Body.String())
			}
		})
	}

	response := handlerRequest(AIAgentPlan, http.MethodPost, "/", `{"task":"`+privateInput+`"}`)
	var payload map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["tool_spec"] == "" || payload["system_prompt"] == "" {
		t.Fatalf("valid request did not preserve the response shape: %#v", payload)
	}
	if transport.requests != 0 {
		t.Fatalf("AI handlers made %d outbound HTTP requests", transport.requests)
	}
}

func TestAIHandlerRejectsMalformedRequestWithoutLeakingDetails(t *testing.T) {
	const detail = `{"question":"`
	response := handlerRequest(AIAnswer, http.MethodPost, "/", detail)
	if response.Code != http.StatusBadRequest || strings.TrimSpace(response.Body.String()) != "bad request" {
		t.Fatalf("got status %d and body %q", response.Code, response.Body.String())
	}
}

func TestSupportExportRedactsSensitiveFields(t *testing.T) {
	var logs bytes.Buffer
	originalLogOutput := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(originalLogOutput)

	response := handlerRequest(SupportExport, http.MethodGet, "/", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	rows, err := csv.NewReader(strings.NewReader(response.Body.String())).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || len(rows[1]) != len(rows[0]) || rows[1][0] != "Support customer" {
		t.Fatalf("valid export shape was not preserved: %#v", rows)
	}
	for _, field := range rows[1][1:] {
		if field != "REDACTED" {
			t.Fatalf("export field was not redacted: %#v", rows[1])
		}
	}
	for _, sensitive := range []string{"123-45-6789", "1978-04-23", "avery@example.invalid"} {
		if strings.Contains(logs.String(), sensitive) {
			t.Fatalf("support export log exposed %q", sensitive)
		}
	}
}

func TestFetchToolValidatesURLAndUsesLocalAdapter(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	attack := handlerRequest(FetchTool, http.MethodGet, "/?url="+url.QueryEscape("http://127.0.0.1/admin"), "")
	if attack.Code != http.StatusBadRequest {
		t.Fatalf("unsafe URL status = %d, want %d", attack.Code, http.StatusBadRequest)
	}

	legitimateURL := "https://tools.example.invalid/reach-testbed-tool.bin"
	response := handlerRequest(FetchTool, http.MethodGet, "/?url="+url.QueryEscape(legitimateURL), "")
	if response.Code != http.StatusOK {
		t.Fatalf("valid URL status = %d, want %d", response.Code, http.StatusOK)
	}
	path := strings.TrimSpace(response.Body.String())
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	if string(content) != "Reachable local tool fixture\n" {
		t.Fatalf("unexpected local tool content %q", content)
	}
}

func TestDiagnosticPingRejectsInjectionAndAllowsHost(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "called")
	ping := filepath.Join(dir, "ping")
	script := "#!/bin/sh\nprintf called >> \"$PING_MARKER\"\nprintf 'PING %s\\n' \"$3\"\n"
	if err := os.WriteFile(ping, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("PING_MARKER", marker)

	attackHost := "example.com;touch /tmp/reachable-command-injection"
	attack := handlerRequest(DiagnosticPing, http.MethodGet, "/?host="+url.QueryEscape(attackHost), "")
	if attack.Code != http.StatusBadRequest {
		t.Fatalf("injection status = %d, want %d", attack.Code, http.StatusBadRequest)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("ping command ran for rejected input: %v", err)
	}
	optionAttack := handlerRequest(SafeDiagnosticPing, http.MethodGet, "/?host=-f", "")
	if optionAttack.Code != http.StatusBadRequest {
		t.Fatalf("ping option status = %d, want %d", optionAttack.Code, http.StatusBadRequest)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("ping command ran for rejected option: %v", err)
	}

	valid := handlerRequest(DiagnosticPing, http.MethodGet, "/?host=example.com", "")
	if valid.Code != http.StatusOK || strings.TrimSpace(valid.Body.String()) != "PING example.com" {
		t.Fatalf("valid host got status %d and body %q", valid.Code, valid.Body.String())
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("ping command did not run for valid input: %v", err)
	}
}

func TestParseLanguageUsesGenericErrorAndAcceptsValidTag(t *testing.T) {
	invalid := handlerRequest(ParseLanguage, http.MethodPost, "/?tag=zz", "")
	if invalid.Code != http.StatusBadRequest || strings.TrimSpace(invalid.Body.String()) != "unsupported language tag" {
		t.Fatalf("invalid tag got status %d and body %q", invalid.Code, invalid.Body.String())
	}
	valid := handlerRequest(ParseLanguage, http.MethodPost, "/?tag=en", "")
	body, err := io.ReadAll(valid.Body)
	if err != nil {
		t.Fatal(err)
	}
	if valid.Code != http.StatusOK || string(body) != "en\n" {
		t.Fatalf("valid tag got status %d and body %q", valid.Code, body)
	}
}
