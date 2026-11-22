package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type ProviderChoice struct {
	ID              int64
	Name, Model     string
	Active, Enabled bool
}

type providerConfigurator interface {
	ConfigureProvider(ctx context.Context, field, value string) (string, error)
	TestActiveProvider(ctx context.Context) (string, error)
}

func (m model) providerCommand(arg string) (tea.Model, tea.Cmd) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return m.providerPicker()
	}
	if m.running || m.busy != "" {
		return m, m.say("Finish or stop the current request first")
	}
	cfg, ok := m.e.(providerConfigurator)
	if !ok {
		return m, m.say("Provider customization is not supported in this engine")
	}
	fields := strings.Fields(arg)
	sub := strings.ToLower(fields[0])
	val := ""
	if len(fields) > 1 {
		val = strings.ToLower(strings.Join(fields[1:], " "))
	}
	switch sub {
	case "test":
		m.busy = "Testing provider"
		return m, tea.Batch(m.spin.Tick, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			out, err := cfg.TestActiveProvider(ctx)
			if err != nil {
				return noticeMsg{err: err}
			}
			return noticeMsg{text: stOK.Render(bullet) + " " + out}
		})
	case "api", "vision", "stream", "failover":
		if val == "" {
			return m, m.say("Usage: /provider api <auto|openai|responses|anthropic> | vision <on|off> | stream <on|off> | failover <on|off> | test")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		out, err := cfg.ConfigureProvider(ctx, sub, val)
		if err != nil {
			m.print(stFail.Render(bullet) + " " + errText(err))
			return m, nil
		}
		m.info = m.e.Info()
		m.print(stOK.Render(bullet) + " " + out)
		return m, nil
	default:
		return m, m.say("Usage: /provider [api|vision|stream|failover|test]")
	}
}

func (m model) providerPicker() (tea.Model, tea.Cmd) {
	if m.running || m.busy != "" {
		return m, m.say("Finish or stop the current request first")
	}
	e, ok := m.e.(interface {
		Providers(context.Context) ([]ProviderChoice, error)
		SetProvider(context.Context, int64) error
	})
	if !ok {
		return m, m.say("Use /connect model to configure a provider")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	providers, err := e.Providers(ctx)
	if err != nil {
		return m, m.say(errText(err))
	}
	if len(providers) == 0 {
		return m, m.say("No provider yet · /connect model")
	}
	labels := []string{}
	ids := map[string]int64{}
	current := ""
	for _, p := range providers {
		label := fmt.Sprintf("%s · %s", p.Name, p.Model)
		if !p.Enabled {
			label += " (off)"
		}
		labels = append(labels, label)
		ids[label] = p.ID
		if p.Active {
			current = label
		}
	}
	m.pick = &picker{title: "Provider", items: labels, current: current, sel: max(indexOf(labels, current), 0), pick: func(label string) tea.Cmd {
		return func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := e.SetProvider(ctx, ids[label]); err != nil {
				return noticeMsg{err: err}
			}
			return noticeMsg{text: "Provider: " + label}
		}
	}}
	return m, nil
}
