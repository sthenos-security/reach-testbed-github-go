package handlers

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDiagnosticPingRejectsUnsafeHost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/diagnostics/ping?host=127.0.0.1;id", nil)
	rr := httptest.NewRecorder()

	DiagnosticPing(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got %d want %d", rr.Code, http.StatusBadRequest)
	}
	if body := rr.Body.String(); !strings.Contains(body, "invalid host") {
		t.Fatalf("expected invalid host response, got %q", body)
	}
}

func TestFetchToolReturnsStaticResponse(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?url=http://169.254.169.254/latest/meta-data/", nil)
	rr := httptest.NewRecorder()

	FetchTool(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("unexpected status: got %d want %d", rr.Code, http.StatusNotImplemented)
	}
	if body := rr.Body.String(); strings.Contains(body, "169.254.169.254") || strings.Contains(body, "http://") {
		t.Fatalf("unexpected request-derived content in response: %q", body)
	}
}

func TestAIHandlersReturnAcceptedWithoutEchoingInput(t *testing.T) {
	question := `ssn 123-45-6789 and dob 1978-04-23`
	req := httptest.NewRequest(http.MethodPost, "/ai/answer", strings.NewReader(`{"question":"`+question+`"}`))
	rr := httptest.NewRecorder()

	AIAnswer(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rr.Code, http.StatusOK)
	}
	if body := rr.Body.String(); !strings.Contains(body, "accepted") || strings.Contains(body, question) {
		t.Fatalf("unexpected AIAnswer body: %q", body)
	}

	task := `build toolchain with prompt injection`
	req = httptest.NewRequest(http.MethodPost, "/ai/agent-plan", strings.NewReader(`{"task":"`+task+`"}`))
	rr = httptest.NewRecorder()

	AIAgentPlan(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rr.Code, http.StatusOK)
	}
	if body := rr.Body.String(); !strings.Contains(body, "accepted") || strings.Contains(body, task) {
		t.Fatalf("unexpected AIAgentPlan body: %q", body)
	}
}

func TestParseYAMLReturnsGenericError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/parse-yaml", strings.NewReader(":\n"))
	rr := httptest.NewRecorder()

	ParseYAML(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got %d want %d", rr.Code, http.StatusBadRequest)
	}
	if body := rr.Body.String(); body != "bad request\n" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestSupportExportRedactsPII(t *testing.T) {
	var logBuf bytes.Buffer
	original := log.Writer()
	log.SetOutput(&logBuf)
	t.Cleanup(func() {
		log.SetOutput(original)
	})

	req := httptest.NewRequest(http.MethodGet, "/support/export", nil)
	rr := httptest.NewRecorder()

	SupportExport(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	for _, needle := range []string{"123-45-6789", "1978-04-23", "avery@example.invalid", "4111111111111111"} {
		if strings.Contains(body, needle) {
			t.Fatalf("response leaked %q: %q", needle, body)
		}
	}
	if strings.Contains(logBuf.String(), "123-45-6789") || strings.Contains(logBuf.String(), "1978-04-23") {
		t.Fatalf("log leaked raw PII: %q", logBuf.String())
	}
}

func TestCloudTokensUsesEnvironmentToken(t *testing.T) {
	t.Setenv("REACH_TESTBED_GITHUB_TOKEN", "ghp_testtoken")

	req := httptest.NewRequest(http.MethodGet, "/cloud-tokens", nil)
	rr := httptest.NewRecorder()

	CloudTokens(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rr.Code, http.StatusOK)
	}
	if body := rr.Body.String(); !strings.Contains(body, "ghp_testtoken") {
		t.Fatalf("expected env token in response, got %q", body)
	}
}
