package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Metixpro/macron-agent/pkg/moa"
	"github.com/Metixpro/macron-agent/pkg/router"
	"github.com/Metixpro/macron-agent/pkg/textutil"
)

func (t *terminal) RoutingCommand(ctx context.Context, arg string) (string, error) {
	arg = strings.TrimSpace(arg)
	rc := t.bot.routerConf(tierOwner)
	if arg == "" || strings.EqualFold(arg, "status") {
		state := "off"
		if rc.on {
			state = "active"
		}
		strongName := "none"
		for _, p := range t.prov.Snapshot() {
			if p.ID == rc.strong {
				strongName = fmt.Sprintf("%s (%s)", p.Name, p.Model)
				break
			}
		}
		return fmt.Sprintf("Smart Routing: %s · strong: %s · sensitivity: %s · escalate: %v", state, strongName, rc.sens, rc.escalate), nil
	}
	fields := strings.Fields(arg)
	sub := strings.ToLower(fields[0])
	switch sub {
	case "on", "off":
		on := sub == "on"
		if err := t.st.SetSetting(ctx, settingRouterOn, strconv.FormatBool(on)); err != nil {
			return "", err
		}
		return "Smart Routing: " + sub, nil
	case "strong":
		if len(fields) < 2 {
			return "", errors.New("usage: /routing strong <provider-name|off>")
		}
		name := strings.Join(fields[1:], " ")
		if strings.EqualFold(name, "off") || strings.EqualFold(name, "none") {
			if err := t.st.SetSetting(ctx, settingRouterStrong+string(tierOwner), ""); err != nil {
				return "", err
			}
			return "Strong model cleared (routing off)", nil
		}
		for _, p := range t.prov.Snapshot() {
			if strings.EqualFold(p.Name, name) || fmt.Sprint(p.ID) == name {
				if err := t.st.SetSetting(ctx, settingRouterStrong+string(tierOwner), strconv.FormatInt(p.ID, 10)); err != nil {
					return "", err
				}
				return fmt.Sprintf("Strong model set to %s (%s)", p.Name, p.Model), nil
			}
		}
		return "", fmt.Errorf("provider %q not found", name)
	case "sensitivity", "sens":
		if len(fields) < 2 {
			return "", errors.New("usage: /routing sensitivity <cheap|balanced|quality>")
		}
		s := router.Sensitivity(strings.ToLower(fields[1]))
		if s != router.Cheap && s != router.Balanced && s != router.Quality {
			return "", errors.New("sensitivity must be cheap, balanced, or quality")
		}
		if err := t.st.SetSetting(ctx, settingRouterSens, string(s)); err != nil {
			return "", err
		}
		return "Routing sensitivity: " + string(s), nil
	default:
		return "", errors.New("usage: /routing [status|on|off|strong <name>|sensitivity <cheap|balanced|quality>]")
	}
}

func (t *terminal) MoaCommand(ctx context.Context, arg string) (string, error) {
	arg = strings.TrimSpace(arg)
	presets := t.bot.moaPresets()
	active, _ := t.st.Setting(settingMoaActive)
	def, _ := t.st.Setting(settingMoaDefault)
	if arg == "" || strings.EqualFold(arg, "status") {
		names := make([]string, 0, len(presets))
		for _, p := range presets {
			names = append(names, p.Name)
		}
		if len(names) == 0 {
			return "MoA: no presets configured yet (create one in Desktop Settings → Smart Routing & MoA, or Telegram /settings)", nil
		}
		return fmt.Sprintf("MoA: active=%q · default=%q · presets: %s", active, def, strings.Join(names, ", ")), nil
	}
	if strings.EqualFold(arg, "off") {
		if err := t.st.SetSetting(ctx, settingMoaActive, ""); err != nil {
			return "", err
		}
		return "MoA always-on mode disabled", nil
	}
	if strings.HasPrefix(strings.ToLower(arg), "on ") {
		name := strings.TrimSpace(arg[3:])
		p, ok := moa.Find(presets, name)
		if !ok {
			return "", fmt.Errorf("MoA preset %q not found", name)
		}
		if probs := t.bot.moaProblems(p); len(probs) > 0 {
			return "", errors.New(strings.Join(probs, "; "))
		}
		if err := t.st.SetSetting(ctx, settingMoaActive, p.Name); err != nil {
			return "", err
		}
		return "MoA active for all turns: " + p.Name, nil
	}
	return "", errors.New("usage: /moa <prompt> | /moa status | /moa on <preset> | /moa off")
}

func (t *terminal) SearchHistory(ctx context.Context, query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "Usage: /search <words> — searches past conversations in this workspace", nil
	}
	hits, err := t.st.SearchArchive(ctx, t.currentUID(), query, time.Time{}, 10)
	if err != nil {
		return "", err
	}
	if len(hits) == 0 {
		return fmt.Sprintf("No matches for %q", query), nil
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Search results for %q (%d):\n", query, len(hits))
	for _, h := range hits {
		who := "user"
		if h.Role == "assistant" {
			who = "macron"
		}
		snip := textutil.TruncateRunes(strings.Join(strings.Fields(h.Snippet), " "), 160, "…")
		fmt.Fprintf(&sb, "  [%s] %s: %s\n", h.CreatedAt.Format("2006-01-02 15:04"), who, snip)
	}
	return strings.TrimRight(sb.String(), "\n"), nil
}
