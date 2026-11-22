package main

import (
	"strings"
	"testing"

	"github.com/Metixpro/macron-agent/pkg/agent"
	"github.com/Metixpro/macron-agent/pkg/app"
)

func TestTelegramCommandMenuMatchesAllBotCommands(t *testing.T) {
	required := []string{
		"start", "help", "status", "queue", "new", "compact",
		"memory", "mode", "files", "cookies", "reminders", "tasks",
		"search", "board", "dashboard", "settings", "execution", "sandbox",
		"update", "learn", "teach", "moa", "approve", "yolo",
		"language", "id", "me", "about", "cancel",
	}

	seen := map[string]commandSpec{}
	for _, spec := range commandSpecs {
		if spec.name == "" || spec.fa == "" || spec.en == "" {
			t.Fatalf("incomplete commandSpec: %+v", spec)
		}
		if _, dup := seen[spec.name]; dup {
			t.Fatalf("duplicate command in commandSpecs: %q", spec.name)
		}
		seen[spec.name] = spec
	}

	for _, name := range required {
		if _, ok := seen[name]; !ok {
			t.Errorf("command /%s is missing from Telegram SetMyCommands (commandSpecs)", name)
		}
		if !strings.Contains(personalHelp, "/"+name) {
			t.Errorf("command /%s is missing from Persian personalHelp", name)
		}
		if enHelp := en.t(personalHelp); !strings.Contains(enHelp, "/"+name) {
			t.Errorf("command /%s is missing from English personalHelp", name)
		}
	}

	ownerMenuFa := commandMenu(fa, commandOwner)
	ownerMenuEn := commandMenu(en, commandOwner)
	if len(ownerMenuFa) != len(commandSpecs) || len(ownerMenuEn) != len(commandSpecs) {
		t.Fatalf("owner command menu mismatch: fa=%d en=%d specs=%d", len(ownerMenuFa), len(ownerMenuEn), len(commandSpecs))
	}
}

func TestSpectorDevIdentityAndLicense(t *testing.T) {
	if app.Creator != "SpectorDev" || app.CreatorURL != "https://t.me/SpectorDev" {
		t.Fatalf("unexpected app creator metadata: %q %q", app.Creator, app.CreatorURL)
	}
	if !strings.Contains(app.License, "MIT") || !strings.Contains(app.License, "SpectorDev") {
		t.Fatalf("unexpected app license: %q", app.License)
	}
	if !strings.Contains(agent.DefaultPersona.Creator, "SpectorDev") || !strings.Contains(agent.DefaultPersona.CreatorHandle, "@SpectorDev") {
		t.Fatalf("DefaultPersona missing SpectorDev identity: %+v", agent.DefaultPersona)
	}
	if !strings.Contains(agent.DefaultSoul, "SpectorDev") || !strings.Contains(agent.DefaultSoul, "MIT") {
		t.Fatalf("DefaultSoul missing SpectorDev/MIT identity: %q", agent.DefaultSoul)
	}
	if !strings.Contains(personalWelcome, "SpectorDev") || !strings.Contains(en.t(personalWelcome), "SpectorDev") {
		t.Fatalf("Telegram welcome missing SpectorDev attribution")
	}
}
