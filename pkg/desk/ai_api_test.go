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

type aiMockHost struct {
	*parityMockHost
	routing RoutingConfig
	moa     MoaState
}

func newAIMockHost(t *testing.T) *aiMockHost {
	t.Helper()
	return &aiMockHost{
		parityMockHost: newParityMockHost(t),
		routing: RoutingConfig{
			Enabled:           true,
			Active:            false,
			LightProviderID:   1,
			LightProviderName: "OpenAI",
			Sensitivity:       "balanced",
			Escalate:          true,
			Kinds:             []string{"code", "reasoning"},
			AllKinds:          []string{"code", "reasoning", "research", "math", "creative"},
		},
		moa: MoaState{
			Presets: []MoaPresetView{},
		},
	}
}

func (m *aiMockHost) RoutingConfig() RoutingConfig {
	cfg := m.routing
	cfg.Active = cfg.Enabled && cfg.StrongProviderID > 0 && cfg.StrongProviderID != cfg.LightProviderID
	return cfg
}

func (m *aiMockHost) SaveRoutingConfig(_ context.Context, req RoutingUpdate) (RoutingConfig, error) {
	if req.Enabled != nil {
		m.routing.Enabled = *req.Enabled
	}
	if req.StrongProviderID != nil {
		m.routing.StrongProviderID = *req.StrongProviderID
		if *req.StrongProviderID == 2 {
			m.routing.StrongProviderName = "Work Claude"
		} else {
			m.routing.StrongProviderName = ""
		}
	}
	if req.Sensitivity != nil {
		m.routing.Sensitivity = *req.Sensitivity
	}
	if req.Escalate != nil {
		m.routing.Escalate = *req.Escalate
	}
	if req.KindsSet {
		m.routing.Kinds = append([]string(nil), req.Kinds...)
	}
	return m.RoutingConfig(), nil
}

func (m *aiMockHost) TestRouting(text string) RoutingTestResult {
	strong := strings.Contains(strings.ToLower(text), "refactor")
	target := m.routing.LightProviderName
	kind := "general"
	if strong {
		kind = "code"
		if m.routing.StrongProviderName != "" {
			target = m.routing.StrongProviderName
		}
	}
	return RoutingTestResult{
		Strong:         strong,
		Kind:           kind,
		Hard:           strong,
		TargetProvider: target,
	}
}

func (m *aiMockHost) MoaState() MoaState { return m.moa }

func (m *aiMockHost) SaveMoaPreset(_ context.Context, p MoaPresetView) (MoaState, error) {
	if strings.TrimSpace(p.Name) == "" {
		return m.moa, errors.New("name required")
	}
	updated := false
	for i := range m.moa.Presets {
		if strings.EqualFold(m.moa.Presets[i].Name, p.Name) {
			m.moa.Presets[i] = p
			updated = true
			break
		}
	}
	if !updated {
		m.moa.Presets = append(m.moa.Presets, p)
	}
	return m.moa, nil
}

func (m *aiMockHost) DeleteMoaPreset(_ context.Context, name string) (MoaState, error) {
	var out []MoaPresetView
	for _, p := range m.moa.Presets {
		if !strings.EqualFold(p.Name, name) {
			out = append(out, p)
		}
	}
	m.moa.Presets = out
	if strings.EqualFold(m.moa.Active, name) {
		m.moa.Active = ""
	}
	return m.moa, nil
}

func (m *aiMockHost) UpdateMoaConfig(_ context.Context, req MoaConfigUpdate) (MoaState, error) {
	if req.Active != nil {
		m.moa.Active = *req.Active
	}
	if req.DefaultPreset != nil {
		m.moa.DefaultPreset = *req.DefaultPreset
	}
	if req.PrivacyFilter != nil {
		m.moa.PrivacyFilter = *req.PrivacyFilter
	}
	return m.moa, nil
}

func doAIReq(t *testing.T, s *Server, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	r := httptest.NewRequest(method, "http://127.0.0.1"+path, &buf)
	r.Host = "127.0.0.1"
	r.AddCookie(&http.Cookie{Name: cookieName, Value: s.token})
	if method != http.MethodGet {
		r.Header.Set("X-Macron", "1")
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}

func TestRoutingAndMoaAPIs(t *testing.T) {
	h := newAIMockHost(t)
	s := New(h)

	// 1. GET /api/routing
	w := doAIReq(t, s, http.MethodGet, "/api/routing", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/routing status=%d body=%s", w.Code, w.Body.String())
	}
	var cfg RoutingConfig
	if err := json.Unmarshal(w.Body.Bytes(), &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Active || cfg.Sensitivity != "balanced" {
		t.Fatalf("unexpected initial routing config: %+v", cfg)
	}

	// 2. POST /api/routing to enable Strong provider
	strongID := int64(2)
	sens := "quality"
	w = doAIReq(t, s, http.MethodPost, "/api/routing", RoutingUpdate{
		StrongProviderID: &strongID,
		Sensitivity:      &sens,
		Kinds:            []string{"code", "reasoning", "research"},
		KindsSet:         true,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("POST /api/routing status=%d body=%s", w.Code, w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &cfg); err != nil {
		t.Fatal(err)
	}
	if !cfg.Active || cfg.StrongProviderID != 2 || cfg.Sensitivity != "quality" || len(cfg.Kinds) != 3 {
		t.Fatalf("unexpected saved routing config: %+v", cfg)
	}

	// 3. POST /api/routing/test
	w = doAIReq(t, s, http.MethodPost, "/api/routing/test", map[string]string{"text": "Refactor the Go HTTP router"})
	if w.Code != http.StatusOK {
		t.Fatalf("POST /api/routing/test status=%d body=%s", w.Code, w.Body.String())
	}
	var tres RoutingTestResult
	if err := json.Unmarshal(w.Body.Bytes(), &tres); err != nil {
		t.Fatal(err)
	}
	if !tres.Strong || tres.Kind != "code" || tres.TargetProvider != "Work Claude" {
		t.Fatalf("unexpected routing test result: %+v", tres)
	}

	// 4. POST /api/moa/presets
	w = doAIReq(t, s, http.MethodPost, "/api/moa/presets", MoaPresetView{
		Name:       "Council",
		Aggregator: MoaSlotView{Provider: "Work Claude"},
		References: []MoaSlotView{{Provider: "OpenAI"}, {Provider: "DeepSeek"}},
		Fanout:     "user_turn",
		Enabled:    true,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("POST /api/moa/presets status=%d body=%s", w.Code, w.Body.String())
	}
	var st MoaState
	if err := json.Unmarshal(w.Body.Bytes(), &st); err != nil {
		t.Fatal(err)
	}
	if len(st.Presets) != 1 || st.Presets[0].Name != "Council" {
		t.Fatalf("unexpected moa presets: %+v", st)
	}

	// 5. POST /api/moa/config (activate + default)
	act := "Council"
	w = doAIReq(t, s, http.MethodPost, "/api/moa/config", MoaConfigUpdate{Active: &act, DefaultPreset: &act})
	if w.Code != http.StatusOK {
		t.Fatalf("POST /api/moa/config status=%d body=%s", w.Code, w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &st); err != nil {
		t.Fatal(err)
	}
	if st.Active != "Council" || st.DefaultPreset != "Council" {
		t.Fatalf("unexpected moa config: %+v", st)
	}

	// 6. DELETE /api/moa/presets/Council
	w = doAIReq(t, s, http.MethodDelete, "/api/moa/presets/Council", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("DELETE /api/moa/presets/Council status=%d body=%s", w.Code, w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &st); err != nil {
		t.Fatal(err)
	}
	if len(st.Presets) != 0 || st.Active != "" {
		t.Fatalf("expected preset deleted and deactivated, got %+v", st)
	}
}

func TestSessionSearchAcrossThreads(t *testing.T) {
	h := newAIMockHost(t)
	s := New(h)

	p := s.project(projectID(h.DefaultProject()))
	if p == nil {
		t.Fatal("default project not initialized")
	}
	p.mu.Lock()
	p.loadThreadsLocked()
	p.loadLocked()
	p.appendLocked(Block{Kind: "user", Text: "Deploy Kubernetes ingress controller on Debian 12", At: 100})
	p.appendLocked(Block{Kind: "assistant", Text: "Here is the Helm chart for your Kubernetes ingress setup.", At: 101})
	p.mu.Unlock()

	// Create a second thread (via /new, archiving the first) and append another message
	w := doAIReq(t, s, http.MethodPost, "/api/projects/"+p.id+"/new", nil)
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Fatalf("create thread status=%d body=%s", w.Code, w.Body.String())
	}
	p.mu.Lock()
	p.appendLocked(Block{Kind: "user", Text: "Optimize PostgreSQL query index performance", At: 200})
	p.mu.Unlock()

	// Search across archived & active threads
	w = doAIReq(t, s, http.MethodGet, "/api/search?q=kubernetes+ingress", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/search status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Hits []SearchHit `json:"hits"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Hits) != 2 {
		t.Fatalf("expected 2 hits across archived thread, got %d: %+v", len(resp.Hits), resp.Hits)
	}
	if resp.Hits[0].Role != "assistant" || resp.Hits[0].ProjectName != "Home" {
		t.Fatalf("unexpected top hit: %+v", resp.Hits[0])
	}

	// Search active thread
	w = doAIReq(t, s, http.MethodGet, "/api/search?q=postgresql", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/search postgresql status=%d", w.Code)
	}
	resp.Hits = nil
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Hits) != 1 || !strings.Contains(resp.Hits[0].Snippet, "PostgreSQL") {
		t.Fatalf("unexpected active thread hit: %+v", resp.Hits)
	}
}
