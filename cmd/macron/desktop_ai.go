package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Metixpro/macron-agent/pkg/desk"
	"github.com/Metixpro/macron-agent/pkg/moa"
	"github.com/Metixpro/macron-agent/pkg/router"
)

func (h *deskHost) providerNameByID(id int64) string {
	if id <= 0 {
		return ""
	}
	for _, s := range h.w.prov.Snapshot() {
		if s.ID == id {
			return s.Name
		}
	}
	return ""
}

func (h *deskHost) RoutingConfig() desk.RoutingConfig {
	rc := h.w.bot.routerConf(tierOwner)
	allKinds := make([]string, len(router.Kinds))
	for i, k := range router.Kinds {
		allKinds[i] = string(k)
	}
	var kinds []string
	for _, k := range router.Kinds {
		if rc.kinds[k] {
			kinds = append(kinds, string(k))
		}
	}
	return desk.RoutingConfig{
		Enabled:            h.w.st.SettingBool(settingRouterOn, true),
		Active:             rc.on,
		LightProviderID:    rc.light,
		LightProviderName:  h.providerNameByID(rc.light),
		StrongProviderID:   rc.strong,
		StrongProviderName: h.providerNameByID(rc.strong),
		Sensitivity:        string(rc.sens),
		Escalate:           rc.escalate,
		Kinds:              kinds,
		AllKinds:           allKinds,
	}
}

func (h *deskHost) SaveRoutingConfig(ctx context.Context, req desk.RoutingUpdate) (desk.RoutingConfig, error) {
	if req.Enabled != nil {
		if err := h.w.st.SetSetting(ctx, settingRouterOn, strconv.FormatBool(*req.Enabled)); err != nil {
			return desk.RoutingConfig{}, err
		}
	}
	if req.StrongProviderID != nil {
		if *req.StrongProviderID < 0 {
			return desk.RoutingConfig{}, errors.New("invalid strong provider id")
		}
		val := ""
		if *req.StrongProviderID > 0 {
			val = strconv.FormatInt(*req.StrongProviderID, 10)
		}
		if err := h.w.st.SetSetting(ctx, settingRouterStrong+string(tierOwner), val); err != nil {
			return desk.RoutingConfig{}, err
		}
	}
	if req.Sensitivity != nil {
		s := router.Sensitivity(strings.ToLower(strings.TrimSpace(*req.Sensitivity)))
		switch s {
		case router.Cheap, router.Balanced, router.Quality:
			if err := h.w.st.SetSetting(ctx, settingRouterSens, string(s)); err != nil {
				return desk.RoutingConfig{}, err
			}
		default:
			return desk.RoutingConfig{}, fmt.Errorf("invalid sensitivity %q (use cheap, balanced, or quality)", *req.Sensitivity)
		}
	}
	if req.Escalate != nil {
		if err := h.w.st.SetSetting(ctx, settingRouterEsc, strconv.FormatBool(*req.Escalate)); err != nil {
			return desk.RoutingConfig{}, err
		}
	}
	if req.KindsSet {
		valid := map[router.Kind]bool{}
		for _, k := range router.Kinds {
			valid[k] = true
		}
		clean := make([]router.Kind, 0, len(req.Kinds))
		for _, raw := range req.Kinds {
			k := router.Kind(strings.ToLower(strings.TrimSpace(raw)))
			if valid[k] {
				clean = append(clean, k)
			}
		}
		b, err := json.Marshal(clean)
		if err != nil {
			return desk.RoutingConfig{}, err
		}
		if err := h.w.st.SetSetting(ctx, settingRouterKinds, string(b)); err != nil {
			return desk.RoutingConfig{}, err
		}
	}
	return h.RoutingConfig(), nil
}

func (h *deskHost) TestRouting(text string) desk.RoutingTestResult {
	rc := h.w.bot.routerConf(tierOwner)
	v := router.Classify(router.Input{Text: text})
	strong := router.Decide(v, rc.kinds, rc.sens)
	sigs := make([]string, 0, len(v.Signals))
	for _, s := range v.Signals {
		sigs = append(sigs, s.String())
	}
	target := h.providerNameByID(rc.light)
	if strong && rc.strong > 0 {
		target = h.providerNameByID(rc.strong)
	}
	return desk.RoutingTestResult{
		Strong:         strong,
		Kind:           string(v.Kind),
		Hard:           v.Hard,
		TargetProvider: target,
		Signals:        sigs,
	}
}

func (h *deskHost) MoaState() desk.MoaState {
	presets := h.w.bot.moaPresets()
	out := make([]desk.MoaPresetView, 0, len(presets))
	for _, p := range presets {
		refs := make([]desk.MoaSlotView, len(p.References))
		for i, r := range p.References {
			refs[i] = desk.MoaSlotView{Provider: r.Provider, Effort: r.Effort}
		}
		out = append(out, desk.MoaPresetView{
			Name:          p.Name,
			Aggregator:    desk.MoaSlotView{Provider: p.Aggregator.Provider, Effort: p.Aggregator.Effort},
			References:    refs,
			Fanout:        p.FanoutName(),
			PrivacyFilter: p.PrivacyFilter,
			Enabled:       p.On(),
			Problems:      h.w.bot.moaProblems(p),
		})
	}
	active, _ := h.w.st.Setting(settingMoaActive)
	def, _ := h.w.st.Setting(settingMoaDefault)
	priv, _ := h.w.st.Setting(settingMoaPrivacy)
	return desk.MoaState{
		Presets:       out,
		Active:        active,
		DefaultPreset: def,
		PrivacyFilter: priv,
	}
}

func (h *deskHost) SaveMoaPreset(ctx context.Context, v desk.MoaPresetView) (desk.MoaState, error) {
	name := strings.TrimSpace(v.Name)
	if name == "" || len([]rune(name)) > 48 {
		return desk.MoaState{}, errors.New("preset name must be 1–48 characters")
	}
	agg := strings.TrimSpace(v.Aggregator.Provider)
	if agg == "" {
		return desk.MoaState{}, errors.New("aggregator provider is required")
	}
	refs := make([]moa.Slot, 0, len(v.References))
	for _, r := range v.References {
		if pr := strings.TrimSpace(r.Provider); pr != "" {
			refs = append(refs, moa.Slot{Provider: pr, Effort: strings.TrimSpace(r.Effort)})
		}
	}
	if len(refs) == 0 {
		return desk.MoaState{}, errors.New("at least one reference provider is required")
	}
	on := true
	preset := moa.Preset{
		Name:          name,
		Aggregator:    moa.Slot{Provider: agg, Effort: strings.TrimSpace(v.Aggregator.Effort)},
		References:    refs,
		Fanout:        strings.TrimSpace(v.Fanout),
		PrivacyFilter: strings.TrimSpace(v.PrivacyFilter),
		Enabled:       &on,
	}
	list := h.w.bot.moaPresets()
	replaced := false
	for i := range list {
		if strings.EqualFold(list[i].Name, name) {
			list[i] = preset
			replaced = true
			break
		}
	}
	if !replaced {
		list = append(list, preset)
	}
	if err := h.w.bot.saveMoaPresets(ctx, list); err != nil {
		return desk.MoaState{}, err
	}
	if def, _ := h.w.st.Setting(settingMoaDefault); def == "" {
		_ = h.w.st.SetSetting(ctx, settingMoaDefault, name)
	}
	return h.MoaState(), nil
}

func (h *deskHost) DeleteMoaPreset(ctx context.Context, name string) (desk.MoaState, error) {
	list := h.w.bot.moaPresets()
	out := make([]moa.Preset, 0, len(list))
	for _, p := range list {
		if !strings.EqualFold(p.Name, name) {
			out = append(out, p)
		}
	}
	if err := h.w.bot.saveMoaPresets(ctx, out); err != nil {
		return desk.MoaState{}, err
	}
	if act, _ := h.w.st.Setting(settingMoaActive); strings.EqualFold(act, name) {
		_ = h.w.st.SetSetting(ctx, settingMoaActive, "")
	}
	if def, _ := h.w.st.Setting(settingMoaDefault); strings.EqualFold(def, name) {
		next := ""
		if len(out) > 0 {
			next = out[0].Name
		}
		_ = h.w.st.SetSetting(ctx, settingMoaDefault, next)
	}
	return h.MoaState(), nil
}

func (h *deskHost) UpdateMoaConfig(ctx context.Context, req desk.MoaConfigUpdate) (desk.MoaState, error) {
	if req.Active != nil {
		val := strings.TrimSpace(*req.Active)
		if val != "" {
			p, ok := moa.Find(h.w.bot.moaPresets(), val)
			if !ok {
				return desk.MoaState{}, fmt.Errorf("unknown MoA preset %q", val)
			}
			if probs := h.w.bot.moaProblems(p); len(probs) > 0 {
				return desk.MoaState{}, errors.New(strings.Join(probs, "; "))
			}
			val = p.Name
		}
		if err := h.w.st.SetSetting(ctx, settingMoaActive, val); err != nil {
			return desk.MoaState{}, err
		}
	}
	if req.DefaultPreset != nil {
		if err := h.w.st.SetSetting(ctx, settingMoaDefault, strings.TrimSpace(*req.DefaultPreset)); err != nil {
			return desk.MoaState{}, err
		}
	}
	if req.PrivacyFilter != nil {
		if err := h.w.st.SetSetting(ctx, settingMoaPrivacy, strings.TrimSpace(*req.PrivacyFilter)); err != nil {
			return desk.MoaState{}, err
		}
	}
	return h.MoaState(), nil
}
