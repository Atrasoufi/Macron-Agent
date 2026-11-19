package main

import (
	"testing"

	"github.com/Metixpro/macron-agent/pkg/app"
	"github.com/Metixpro/macron-agent/pkg/config"
	"github.com/Metixpro/macron-agent/pkg/machine"
)

func TestApplySavedExecutionMode_LoadsFromStoreWhenEnvUnset(t *testing.T) {
	t.Setenv(envExecutionMode, "")

	cfg := &config.Config{ExecutionMode: "best-effort"}
	applySavedExecutionMode(cfg, func(key string) (string, bool) {
		if key != settingExecutionMode {
			t.Fatalf("unexpected setting key %q", key)
		}
		return "sandbox", true
	})

	if cfg.ExecutionMode != "sandbox" {
		t.Fatalf("expected saved mode %q, got %q", "sandbox", cfg.ExecutionMode)
	}
}

func TestApplySavedExecutionMode_EnvOverridesSavedSetting(t *testing.T) {
	t.Setenv(envExecutionMode, "sandbox")

	called := false
	cfg := &config.Config{ExecutionMode: "sandbox"}
	applySavedExecutionMode(cfg, func(key string) (string, bool) {
		called = true
		return "local", true
	})

	if called {
		t.Fatalf("expected saved setting lookup to be skipped when %s is set", envExecutionMode)
	}
	if cfg.ExecutionMode != "sandbox" {
		t.Fatalf("expected env mode %q to win over saved setting, got %q", "sandbox", cfg.ExecutionMode)
	}
}

func TestApplySavedExecutionMode_IgnoresInvalidOrMissingSetting(t *testing.T) {
	t.Setenv(envExecutionMode, "")

	cfg := &config.Config{ExecutionMode: "best-effort"}
	applySavedExecutionMode(cfg, func(string) (string, bool) {
		return "corrupted-value", true
	})
	if cfg.ExecutionMode != "best-effort" {
		t.Fatalf("expected invalid saved mode to be ignored, got %q", cfg.ExecutionMode)
	}

	applySavedExecutionMode(cfg, func(string) (string, bool) {
		return "", false
	})
	if cfg.ExecutionMode != "best-effort" {
		t.Fatalf("expected missing setting to keep default mode, got %q", cfg.ExecutionMode)
	}
}

func TestTerminalExecutionStatus(t *testing.T) {
	t.Setenv(envExecutionMode, "")
	t.Setenv("MACRON_SANDBOX", "")

	m, err := machine.New(machine.Config{
		BaseDir:       t.TempDir(),
		Sandbox:       machine.SandboxOff,
		ExecutionMode: "local",
	})
	if err != nil {
		t.Fatalf("unexpected machine init error: %v", err)
	}

	term := &terminal{app: &app.App{Machine: m}}
	if got := term.executionStatus(); got != "local (unsandboxed)" {
		t.Fatalf("expected %q, got %q", "local (unsandboxed)", got)
	}
}