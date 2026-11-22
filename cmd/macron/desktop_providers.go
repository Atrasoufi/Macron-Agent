package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Metixpro/macron-agent/pkg/agent"
	"github.com/Metixpro/macron-agent/pkg/brain"
	"github.com/Metixpro/macron-agent/pkg/desk"
	"github.com/Metixpro/macron-agent/pkg/llm"
	"github.com/Metixpro/macron-agent/pkg/providers"
	"github.com/Metixpro/macron-agent/pkg/store"
)

func deskMaskKey(k string) string {
	switch {
	case k == "":
		return ""
	case len(k) <= 8:
		return "••••"
	}
	return k[:3] + "…" + k[len(k)-4:]
}

func (h *deskHost) toProviderDetail(p store.Provider, active, cooling bool) desk.ProviderDetail {
	resolved := providers.API(p)
	limit := 2.0
	if resolved == llm.APIAnthropic {
		limit = 1.0
	}
	d := desk.ProviderDetail{
		ID:            p.ID,
		Name:          p.Name,
		BaseURL:       p.BaseURL,
		APIKeySet:     p.APIKey != "",
		APIKeyMasked:  deskMaskKey(p.APIKey),
		Model:         p.Model,
		API:           p.API,
		ResolvedAPI:   resolved,
		Priority:      p.Priority,
		Enabled:       p.Enabled,
		Active:        active,
		Vision:        p.Vision,
		NoStream:      p.NoStream,
		ContextWindow: p.ContextWindow,
		AutoWindow:    agent.ContextWindow(p.Model),
		Temperature:   p.Temperature,
		TempLimit:     limit,
		Effort:        p.Effort,
		Requests:      p.Requests,
		Failures:      p.Failures,
		AvgLatencyMS:  p.AvgLatencyMS,
		LastError:     p.LastError,
		Cooling:       cooling,
	}
	if p.LastOKAt != nil {
		d.LastOKAt = p.LastOKAt.UTC().Format(time.RFC3339)
	}
	if p.LastErrorAt != nil {
		d.LastErrorAt = p.LastErrorAt.UTC().Format(time.RFC3339)
	}
	return d
}

func (h *deskHost) ProviderDetails(ctx context.Context) ([]desk.ProviderDetail, error) {
	if err := h.w.prov.Reload(ctx); err != nil {
		return nil, err
	}
	snap := h.w.prov.Snapshot()
	out := make([]desk.ProviderDetail, 0, len(snap))
	for _, s := range snap {
		out = append(out, h.toProviderDetail(s.Provider, s.Active, s.Cooling))
	}
	return out, nil
}

func validateProviderAPI(api string) (string, error) {
	api = strings.ToLower(strings.TrimSpace(api))
	switch api {
	case "", "auto":
		return "", nil
	case llm.APIOpenAI, llm.APIResponses, llm.APIAnthropic:
		return api, nil
	default:
		return "", fmt.Errorf("unsupported API format %q (use auto, openai, responses, or anthropic)", api)
	}
}

func validateProviderEffort(eff string) (string, error) {
	eff = strings.ToLower(strings.TrimSpace(eff))
	switch eff {
	case "", "auto", "off", "low", "medium", "high", "xhigh", "ultracode":
		if eff == "auto" {
			return "", nil
		}
		return eff, nil
	default:
		return "", fmt.Errorf("unsupported effort %q", eff)
	}
}

func (h *deskHost) AddProvider(ctx context.Context, req desk.ProviderUpsert) (desk.ProviderDetail, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || len([]rune(name)) > 48 {
		return desk.ProviderDetail{}, errors.New("provider name must be 1–48 characters")
	}
	base, inferredAPI := llm.CleanBaseURL(strings.TrimSpace(req.BaseURL))
	api, err := validateProviderAPI(req.API)
	if err != nil {
		return desk.ProviderDetail{}, err
	}
	if api == "" && inferredAPI != "" {
		api = inferredAPI
	}
	if err := providers.ValidateBaseURL(base); err != nil {
		return desk.ProviderDetail{}, err
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		return desk.ProviderDetail{}, errors.New("a model name is required")
	}

	nextPrio := 10
	if list, err := h.w.st.Providers(ctx); err == nil {
		for _, x := range list {
			if x.Priority+10 > nextPrio {
				nextPrio = x.Priority + 10
			}
		}
	}

	p := store.Provider{
		Name:     name,
		BaseURL:  base,
		APIKey:   strings.TrimSpace(req.APIKey),
		Model:    model,
		API:      api,
		Priority: nextPrio,
		Enabled:  true,
		Vision:   true,
	}
	if req.Priority != nil && *req.Priority >= 0 {
		p.Priority = *req.Priority
	}
	if req.Enabled != nil {
		p.Enabled = *req.Enabled
	}
	if req.Vision != nil {
		p.Vision = *req.Vision
	}
	if req.NoStream != nil {
		p.NoStream = *req.NoStream
	}
	if req.ContextWindow != nil {
		if *req.ContextWindow < 0 || (*req.ContextWindow > 0 && (*req.ContextWindow < 8000 || *req.ContextWindow > 10_000_000)) {
			return desk.ProviderDetail{}, errors.New("context window must be 0 (auto) or between 8,000 and 10,000,000 tokens")
		}
		p.ContextWindow = *req.ContextWindow
	}
	if req.Effort != nil {
		eff, err := validateProviderEffort(*req.Effort)
		if err != nil {
			return desk.ProviderDetail{}, err
		}
		p.Effort = eff
	}
	if !req.ClearTemp && req.Temperature != nil {
		maxT := 2.0
		if providers.API(p) == llm.APIAnthropic {
			maxT = 1.0
		}
		if *req.Temperature < 0 || *req.Temperature > maxT {
			return desk.ProviderDetail{}, fmt.Errorf("temperature must be between 0 and %.1f", maxT)
		}
		v := *req.Temperature
		p.Temperature = &v
	}

	p, err = providers.Normalize(p)
	if err != nil {
		return desk.ProviderDetail{}, err
	}
	if !req.SkipTest {
		if _, err := h.w.prov.Test(ctx, p); err != nil {
			return desk.ProviderDetail{}, fmt.Errorf("connection test failed: %w", err)
		}
	}
	id, err := h.w.st.AddProvider(ctx, p)
	if errors.Is(err, store.ErrDuplicate) {
		return desk.ProviderDetail{}, fmt.Errorf("a provider named %q already exists", p.Name)
	}
	if err != nil {
		return desk.ProviderDetail{}, err
	}
	p.ID = id
	if req.Activate || !h.w.prov.Configured() {
		if err := h.w.st.SetSetting(ctx, store.SettingActiveProvider, strconv.FormatInt(id, 10)); err != nil {
			return desk.ProviderDetail{}, err
		}
	}
	if err := h.w.prov.Reload(ctx); err != nil {
		return desk.ProviderDetail{}, err
	}
	return h.toProviderDetail(p, h.w.prov.ActiveID() == id, false), nil
}

func (h *deskHost) UpdateProvider(ctx context.Context, id int64, req desk.ProviderUpsert) (desk.ProviderDetail, error) {
	p, err := h.w.st.Provider(ctx, id)
	if err != nil {
		return desk.ProviderDetail{}, err
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		if len([]rune(name)) > 48 {
			return desk.ProviderDetail{}, errors.New("provider name must be 48 characters or less")
		}
		p.Name = name
	}
	if rawURL := strings.TrimSpace(req.BaseURL); rawURL != "" {
		base, inferred := llm.CleanBaseURL(rawURL)
		if err := providers.ValidateBaseURL(base); err != nil {
			return desk.ProviderDetail{}, err
		}
		p.BaseURL = base
		if req.API == "" && inferred != "" {
			p.API = inferred
		}
	}
	if req.API != "" {
		api, err := validateProviderAPI(req.API)
		if err != nil {
			return desk.ProviderDetail{}, err
		}
		p.API = api
	}
	if req.ClearKey {
		p.APIKey = ""
	} else if key := strings.TrimSpace(req.APIKey); key != "" {
		p.APIKey = key
	}
	if model := strings.TrimSpace(req.Model); model != "" {
		p.Model = model
	}
	if req.Priority != nil && *req.Priority >= 0 {
		p.Priority = *req.Priority
	}
	if req.Enabled != nil {
		p.Enabled = *req.Enabled
	}
	if req.Vision != nil {
		p.Vision = *req.Vision
	}
	if req.NoStream != nil {
		p.NoStream = *req.NoStream
	}
	if req.ContextWindow != nil {
		if *req.ContextWindow < 0 || (*req.ContextWindow > 0 && (*req.ContextWindow < 8000 || *req.ContextWindow > 10_000_000)) {
			return desk.ProviderDetail{}, errors.New("context window must be 0 (auto) or between 8,000 and 10,000,000 tokens")
		}
		p.ContextWindow = *req.ContextWindow
	}
	if req.Effort != nil {
		eff, err := validateProviderEffort(*req.Effort)
		if err != nil {
			return desk.ProviderDetail{}, err
		}
		p.Effort = eff
	}
	if req.ClearTemp {
		p.Temperature = nil
	} else if req.Temperature != nil {
		maxT := 2.0
		if providers.API(*p) == llm.APIAnthropic {
			maxT = 1.0
		}
		if *req.Temperature < 0 || *req.Temperature > maxT {
			return desk.ProviderDetail{}, fmt.Errorf("temperature must be between 0 and %.1f", maxT)
		}
		v := *req.Temperature
		p.Temperature = &v
	}

	normalized, err := providers.Normalize(*p)
	if err != nil {
		return desk.ProviderDetail{}, err
	}
	if err := h.w.st.UpdateProvider(ctx, normalized); err != nil {
		return desk.ProviderDetail{}, err
	}
	if req.Activate && normalized.Enabled {
		_ = h.w.st.SetSetting(ctx, store.SettingActiveProvider, strconv.FormatInt(normalized.ID, 10))
	}
	if err := h.w.prov.Reload(ctx); err != nil {
		return desk.ProviderDetail{}, err
	}
	return h.toProviderDetail(normalized, h.w.prov.ActiveID() == normalized.ID, false), nil
}

func (h *deskHost) DeleteProvider(ctx context.Context, id int64) error {
	if _, err := h.w.st.Provider(ctx, id); err != nil {
		return err
	}
	if err := h.w.st.DeleteProvider(ctx, id); err != nil {
		return err
	}
	_ = h.w.st.SetSetting(ctx, fmt.Sprintf("provider_tools:%d", id), "")
	if h.w.prov.ActiveID() == id {
		nextID := ""
		if list, err := h.w.st.Providers(ctx); err == nil {
			for _, x := range list {
				if x.Enabled {
					nextID = strconv.FormatInt(x.ID, 10)
					break
				}
			}
		}
		_ = h.w.st.SetSetting(ctx, store.SettingActiveProvider, nextID)
	}
	return h.w.prov.Reload(ctx)
}

func (h *deskHost) TestProvider(ctx context.Context, id int64) (desk.ProviderTestResult, error) {
	p, err := h.w.st.Provider(ctx, id)
	if err != nil {
		return desk.ProviderTestResult{}, err
	}
	res, err := h.w.prov.Test(ctx, *p)
	if err != nil {
		return desk.ProviderTestResult{OK: false, Error: err.Error()}, nil
	}
	return desk.ProviderTestResult{
		OK:        true,
		Reply:     shortText(res.Reply, 120),
		LatencyMS: res.Latency.Milliseconds(),
	}, nil
}

func (h *deskHost) ProviderModels(ctx context.Context, id int64) ([]string, error) {
	p, err := h.w.st.Provider(ctx, id)
	if err != nil {
		return nil, err
	}
	return h.w.prov.Models(ctx, *p)
}

func (h *deskHost) Failover() bool {
	return h.w.st.SettingBool(store.SettingFailover, true)
}

func (h *deskHost) SetFailover(ctx context.Context, on bool) error {
	return h.w.st.SetSetting(ctx, store.SettingFailover, strconv.FormatBool(on))
}

func (h *deskHost) Soul() (string, string) {
	if h.w.bot == nil || h.w.bot.brain == nil {
		return "", ""
	}
	return h.w.bot.brain.SoulText(), filepath.Join(h.w.bot.brain.Dir, brain.SoulFile)
}

func (h *deskHost) SaveSoul(text string) error {
	if h.w.bot == nil || h.w.bot.brain == nil {
		return errors.New("SOUL store is not initialized")
	}
	return h.w.bot.brain.SetSoul(text)
}

func (h *deskHost) ResetSoul() (string, error) {
	if h.w.bot == nil || h.w.bot.brain == nil {
		return "", errors.New("SOUL store is not initialized")
	}
	if err := h.w.bot.brain.ResetSoul(); err != nil {
		return "", err
	}
	return h.w.bot.brain.SoulText(), nil
}
