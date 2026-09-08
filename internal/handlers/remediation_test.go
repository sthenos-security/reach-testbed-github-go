package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestAIAnswerRejectsEmptyAndKeepsLocalPrompt(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/ai/answer", strings.NewReader(`{"question":""}`))
	rec := httptest.NewRecorder()
	AIAnswer(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty question: status=%d body=%q", rec.Code, rec.Body.String())
	}

	body := `{"question":"What is my PTO balance?"}`
	req = httptest.NewRequest(http.MethodPost, "/ai/answer", strings.NewReader(body))
	rec = httptest.NewRecorder()
	AIAnswer(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("legitimate question: status=%d body=%q", rec.Code, rec.Body.String())
	}
	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.Contains(got["prompt"], "PTO balance") {
		t.Fatalf("expected quoted user data in prompt, got %#v", got)
	}
	if strings.Contains(rec.Body.String(), "api.openai.com") {
		t.Fatal("response must not reference outbound LLM endpoint")
	}
}

func TestAIAgentPlanIsolatesTaskWithoutOutbound(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/ai/agent-plan", strings.NewReader(`{"task":""}`))
	rec := httptest.NewRecorder()
	AIAgentPlan(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty task: status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/ai/agent-plan", strings.NewReader(`{"task":"summarize ticket 42"}`))
	rec = httptest.NewRecorder()
	AIAgentPlan(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("legitimate task: status=%d body=%q", rec.Code, rec.Body.String())
	}
	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["system_prompt"] == "" || !strings.Contains(got["tool_spec"], "ticket 42") {
		t.Fatalf("unexpected plan payload %#v", got)
	}
}

func TestAIAnswerDecodeErrorIsGeneric(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/ai/answer", strings.NewReader("{not-json"))
	rec := httptest.NewRecorder()
	AIAnswer(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "invalid character") {
		t.Fatalf("leaked parser detail: %q", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "bad request") {
		t.Fatalf("expected generic error, got %q", rec.Body.String())
	}
}

func TestSupportExportOmitsPIIFromLogsAndKeepsCSV(t *testing.T) {
	var logBuf bytes.Buffer
	prev := testingLogWriter(&logBuf)
	defer prev()

	req := httptest.NewRequest(http.MethodGet, "/support/export", nil)
	rec := httptest.NewRecorder()
	SupportExport(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Avery Example") {
		t.Fatalf("expected CSV export body, got %q", body)
	}
	logged := logBuf.String()
	if strings.Contains(logged, "123-45-6789") || strings.Contains(logged, "1978-04-23") {
		t.Fatalf("log leaked PII: %q", logged)
	}
}

func TestFetchToolRejectsSSRFAndHTTPAndAcceptsAllowlisted(t *testing.T) {
	attack := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?"+url.Values{"url": {"http://127.0.0.1/secret"}}.Encode(), nil)
	rec := httptest.NewRecorder()
	FetchTool(rec, attack)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("http SSRF candidate: status=%d body=%q", rec.Code, rec.Body.String())
	}
	if strings.Contains(strings.ToLower(rec.Body.String()), "connection refused") {
		t.Fatalf("leaked dial error: %q", rec.Body.String())
	}

	unlisted := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?"+url.Values{"url": {"https://evil.example/tool"}}.Encode(), nil)
	rec = httptest.NewRecorder()
	FetchTool(rec, unlisted)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unlisted host: status=%d body=%q", rec.Code, rec.Body.String())
	}

	okReq := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?"+url.Values{"url": {"https://example.invalid/tool.bin"}}.Encode(), nil)
	rec = httptest.NewRecorder()
	FetchTool(rec, okReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("allowlisted: status=%d body=%q", rec.Code, rec.Body.String())
	}
	staged := strings.TrimSpace(rec.Body.String())
	data, err := os.ReadFile(staged)
	if err != nil {
		t.Fatalf("read staged tool: %v", err)
	}
	if !strings.Contains(string(data), "https://example.invalid/tool.bin") {
		t.Fatalf("staged fixture missing static URL, got %q", data)
	}
	_ = os.Remove(staged)
}

func TestDiagnosticPingRejectsInjectionAndAcceptsHostname(t *testing.T) {
	attack := httptest.NewRequest(http.MethodGet, "/diagnostics/ping?"+url.Values{"host": {"127.0.0.1; id"}}.Encode(), nil)
	rec := httptest.NewRecorder()
	DiagnosticPing(rec, attack)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("injection host: status=%d body=%q", rec.Code, rec.Body.String())
	}

	okReq := httptest.NewRequest(http.MethodGet, "/diagnostics/ping?"+url.Values{"host": {"localhost"}}.Encode(), nil)
	rec = httptest.NewRecorder()
	DiagnosticPing(rec, okReq)
	if rec.Code != http.StatusOK && rec.Code != http.StatusBadGateway {
		t.Fatalf("legitimate host unexpected status=%d body=%q", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "sh: ") {
		t.Fatalf("shell error leaked: %q", rec.Body.String())
	}
}

func TestCloudTokensDoesNotEchoGitHubSecret(t *testing.T) {
	t.Setenv("REACH_TESTBED_GITHUB_TOKEN", "runtime-test-value")
	req := httptest.NewRequest(http.MethodGet, "/cloud-tokens", nil)
	rec := httptest.NewRecorder()
	CloudTokens(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "runtime-test-value") || strings.Contains(body, "ghp_") {
		t.Fatalf("secret echoed: %q", body)
	}
	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["github_token"] != "configured" {
		t.Fatalf("expected configured status, got %#v", got)
	}

	t.Setenv("REACH_TESTBED_GITHUB_TOKEN", "")
	rec = httptest.NewRecorder()
	CloudTokens(rec, req)
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode empty: %v", err)
	}
	if got["github_token"] != "not-configured" {
		t.Fatalf("expected not-configured, got %#v", got)
	}
}

func TestParseYAMLAndLanguageGenericErrors(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/parse-yaml", strings.NewReader(": not yaml"))
	rec := httptest.NewRecorder()
	ParseYAML(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("yaml status=%d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "yaml:") || strings.Contains(rec.Body.String(), "unmarshal") {
		t.Fatalf("yaml error leaked: %q", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/parse-language?"+url.Values{"tag": {"en"}}.Encode(), nil)
	rec = httptest.NewRecorder()
	ParseLanguage(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("language status=%d body=%q", rec.Code, rec.Body.String())
	}
	got, _ := io.ReadAll(io.LimitReader(rec.Body, 64))
	if strings.TrimSpace(string(got)) == "" {
		t.Fatal("expected language tag in response")
	}
}

// testingLogWriter redirects the default log package output used by handlers.
func testingLogWriter(buf *bytes.Buffer) func() {
	prev := log.Writer()
	flags := log.Flags()
	prefix := log.Prefix()
	log.SetOutput(buf)
	return func() {
		log.SetOutput(prev)
		log.SetFlags(flags)
		log.SetPrefix(prefix)
	}
}
