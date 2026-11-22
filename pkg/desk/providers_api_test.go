package desk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type parityMockHost struct {
	*testFeaturesHost
	providers map[int64]ProviderDetail
	nextID    int64
	activeID  int64
	failover  bool
	soul      string
}

func newParityMockHost(t *testing.T) *parityMockHost {
	t.Helper()
	return &parityMockHost{
		testFeaturesHost: &testFeaturesHost{
			dataDir: t.TempDir(),
			projDir: t.TempDir(),
		},
		providers: map[int64]ProviderDetail{
			1: {
				ID: 1, Name: "OpenAI", BaseURL: "https://api.openai.com/v1", APIKeySet: true, APIKeyMasked: "sk-…1234",
				Model: "gpt-4o", API: "openai", ResolvedAPI: "openai", Priority: 10, Enabled: true, Active: true, Vision: true,
			},
		},
		nextID:   2,
		activeID: 1,
		failover: true,
		soul:     "Default SpectorDev soul",
	}
}

func (m *parityMockHost) ProviderDetails(context.Context) ([]ProviderDetail, error) {
	out := make([]ProviderDetail, 0, len(m.providers))
	for _, p := range m.providers {
		p.Active = p.ID == m.activeID
		out = append(out, p)
	}
	return out, nil
}

func (m *parityMockHost) AddProvider(_ context.Context, req ProviderUpsert) (ProviderDetail, error) {
	id := m.nextID
	m.nextID++
	p := ProviderDetail{
		ID: id, Name: req.Name, BaseURL: req.BaseURL, APIKeySet: req.APIKey != "",
		Model: req.Model, API: req.API, ResolvedAPI: req.API, Enabled: true, Vision: true,
	}
	if req.Vision != nil {
		p.Vision = *req.Vision
	}
	if req.NoStream != nil {
		p.NoStream = *req.NoStream
	}
	if req.ContextWindow != nil {
		p.ContextWindow = *req.ContextWindow
	}
	if req.Temperature != nil {
		v := *req.Temperature
		p.Temperature = &v
	}
	if req.Effort != nil {
		p.Effort = *req.Effort
	}
	m.providers[id] = p
	if req.Activate {
		m.activeID = id
		p.Active = true
	}
	return p, nil
}

func (m *parityMockHost) UpdateProvider(_ context.Context, id int64, req ProviderUpsert) (ProviderDetail, error) {
	p, ok := m.providers[id]
	if !ok {
		return ProviderDetail{}, errors.New("not found")
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Model != "" {
		p.Model = req.Model
	}
	if req.API != "" {
		p.API = req.API
		p.ResolvedAPI = req.API
	}
	if req.Vision != nil {
		p.Vision = *req.Vision
	}
	if req.NoStream != nil {
		p.NoStream = *req.NoStream
	}
	if req.ContextWindow != nil {
		p.ContextWindow = *req.ContextWindow
	}
	if req.ClearTemp {
		p.Temperature = nil
	} else if req.Temperature != nil {
		v := *req.Temperature
		p.Temperature = &v
	}
	if req.Effort != nil {
		p.Effort = *req.Effort
	}
	if req.Enabled != nil {
		p.Enabled = *req.Enabled
	}
	m.providers[id] = p
	p.Active = p.ID == m.activeID
	return p, nil
}

func (m *parityMockHost) DeleteProvider(_ context.Context, id int64) error {
	delete(m.providers, id)
	return nil
}

func (m *parityMockHost) TestProvider(_ context.Context, id int64) (ProviderTestResult, error) {
	if _, ok := m.providers[id]; !ok {
		return ProviderTestResult{}, errors.New("not found")
	}
	return ProviderTestResult{OK: true, Reply: "pong", LatencyMS: 88}, nil
}

func (m *parityMockHost) ProviderModels(_ context.Context, id int64) ([]string, error) {
	if _, ok := m.providers[id]; !ok {
		return nil, errors.New("not found")
	}
	return []string{"custom-model-a", "custom-model-b"}, nil
}

func (m *parityMockHost) Failover() bool { return m.failover }
func (m *parityMockHost) SetFailover(_ context.Context, on bool) error {
	m.failover = on
	return nil
}

func (m *parityMockHost) Soul() (string, string) { return m.soul, "/tmp/SOUL.md" }
func (m *parityMockHost) SaveSoul(text string) error {
	m.soul = text
	return nil
}
func (m *parityMockHost) ResetSoul() (string, error) {
	m.soul = "Default SpectorDev soul"
	return m.soul, nil
}

func (s *fakeSession) Shell(_ context.Context, command string) string {
	return "$ " + command + "\nok-shell-output\n[exit 0 · 2ms]"
}

func (s *fakeSession) Status(_ context.Context) string {
	return "Macron Agent 1.17.0 — by SpectorDev (https://t.me/SpectorDev) · MIT License"
}

func TestDesktopProviderCRUDFailoverSoulShellAndStatus(t *testing.T) {
	h := newParityMockHost(t)
	srv := New(h)
	handler := srv.Handler()
	req := func(method, path string, body any) *httptest.ResponseRecorder {
		var buf bytes.Buffer
		if body != nil {
			_ = json.NewEncoder(&buf).Encode(body)
		}
		r := httptest.NewRequest(method, "http://127.0.0.1"+path, &buf)
		r.Host = "127.0.0.1"
		r.Header.Set("X-Macron", "1")
		r.AddCookie(&http.Cookie{Name: cookieName, Value: srv.token})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}

	// 1. Add custom named provider with per-model settings
	vis := false
	noStream := true
	win := 256000
	temp := 0.35
	eff := "high"
	w := req("POST", "/api/providers", ProviderUpsert{
		Name: "My DeepSeek", BaseURL: "https://api.deepseek.com/v1", APIKey: "sk-test",
		Model: "deepseek-reasoner", API: "openai", Vision: &vis, NoStream: &noStream,
		ContextWindow: &win, Temperature: &temp, Effort: &eff, SkipTest: true, Activate: true,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("add provider failed (%d): %s", w.Code, w.Body.String())
	}
	var addResp struct {
		Provider ProviderDetail `json:"provider"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &addResp)
	if addResp.Provider.Name != "My DeepSeek" || addResp.Provider.ContextWindow != 256000 || !addResp.Provider.NoStream || addResp.Provider.Vision {
		t.Fatalf("unexpected added provider: %+v", addResp.Provider)
	}

	// 2. Update provider
	newWin := 400000
	w = req("PUT", "/api/providers/2", ProviderUpsert{API: "responses", ContextWindow: &newWin})
	if w.Code != http.StatusOK {
		t.Fatalf("update provider failed (%d): %s", w.Code, w.Body.String())
	}

	// 3. Test & list models for saved provider
	w = req("POST", "/api/providers/2/test", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"latencyMs":88`) {
		t.Fatalf("test provider failed (%d): %s", w.Code, w.Body.String())
	}
	w = req("GET", "/api/providers/2/models", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "custom-model-a") {
		t.Fatalf("provider models failed (%d): %s", w.Code, w.Body.String())
	}

	// 4. Toggle failover
	w = req("POST", "/api/failover", map[string]bool{"on": false})
	if w.Code != http.StatusOK || h.failover {
		t.Fatalf("failover toggle failed (%d): %s", w.Code, w.Body.String())
	}

	// 5. Delete provider
	w = req("DELETE", "/api/providers/2", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete provider failed (%d): %s", w.Code, w.Body.String())
	}

	// 6. SOUL editor save + reset
	w = req("POST", "/api/soul", map[string]string{"text": "Custom SpectorDev persona"})
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Custom SpectorDev persona") {
		t.Fatalf("save soul failed (%d): %s", w.Code, w.Body.String())
	}
	w = req("POST", "/api/soul", map[string]bool{"reset": true})
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Default SpectorDev soul") {
		t.Fatalf("reset soul failed (%d): %s", w.Code, w.Body.String())
	}

	// 7. Inline !shell and /status on default project
	w = req("GET", "/api/state", nil)
	var st struct {
		Default string `json:"default"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &st)
	w = req("POST", "/api/projects/"+st.Default+"/shell", map[string]string{"command": "echo hi"})
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "ok-shell-output") {
		t.Fatalf("shell api failed (%d): %s", w.Code, w.Body.String())
	}
	w = req("GET", "/api/projects/"+st.Default+"/status", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "SpectorDev") {
		t.Fatalf("status api failed (%d): %s", w.Code, w.Body.String())
	}
}
