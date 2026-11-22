package llm

import "testing"

func TestEstimateCost(t *testing.T) {
	usd, ok := EstimateCost("openai/gpt-4o-mini", Usage{PromptTokens: 1_000_000, CompletionTokens: 1_000_000})
	if !ok {
		t.Fatal("expected known price for gpt-4o-mini")
	}
	if usd < 0.74 || usd > 0.76 {
		t.Fatalf("unexpected gpt-4o-mini cost: %f", usd)
	}

	usd2, ok := EstimateCost("claude-sonnet-4-5", Usage{PromptTokens: 100_000, CachedTokens: 50_000, CompletionTokens: 10_000})
	if !ok || usd2 <= 0 {
		t.Fatalf("expected positive claude-sonnet-4-5 cost, got %f (ok=%v)", usd2, ok)
	}

	if _, ok := EstimateCost("my-local-unknown-gguf", Usage{PromptTokens: 1000}); ok {
		t.Fatal("expected unknown price for unrecognized local model")
	}

	if got := FormatCost(0.0042); got != "$0.0042" {
		t.Fatalf("FormatCost(0.0042) = %q", got)
	}
	if got := FormatCost(1.256); got != "$1.26" {
		t.Fatalf("FormatCost(1.256) = %q", got)
	}
}
