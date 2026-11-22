package llm

import (
	"fmt"
	"strings"
)

// ModelPrice holds approximate USD cost per 1,000,000 tokens.
type ModelPrice struct {
	InputPerM  float64
	OutputPerM float64
	CachePerM  float64
}

type priceRule struct {
	match string
	price ModelPrice
}

// Ordered from most specific to most general so e.g. "gpt-4o-mini" matches before "gpt-4o".
var priceTable = []priceRule{
	// OpenAI
	{"gpt-5-nano", ModelPrice{0.05, 0.40, 0.005}},
	{"gpt-5-mini", ModelPrice{0.25, 2.00, 0.025}},
	{"gpt-5", ModelPrice{1.25, 10.00, 0.125}},
	{"gpt-4.1-nano", ModelPrice{0.10, 0.40, 0.025}},
	{"gpt-4.1-mini", ModelPrice{0.40, 1.60, 0.10}},
	{"gpt-4.1", ModelPrice{2.00, 8.00, 0.50}},
	{"gpt-4o-mini", ModelPrice{0.15, 0.60, 0.075}},
	{"gpt-4o", ModelPrice{2.50, 10.00, 1.25}},
	{"o4-mini", ModelPrice{1.10, 4.40, 0.275}},
	{"o3-mini", ModelPrice{1.10, 4.40, 0.55}},
	{"o3-pro", ModelPrice{20.00, 80.00, 10.00}},
	{"o3", ModelPrice{2.00, 8.00, 0.50}},
	{"o1-mini", ModelPrice{1.10, 4.40, 0.55}},
	{"o1", ModelPrice{15.00, 60.00, 7.50}},

	// Anthropic Claude
	{"claude-opus-4", ModelPrice{15.00, 75.00, 1.50}},
	{"claude-sonnet-4", ModelPrice{3.00, 15.00, 0.30}},
	{"claude-haiku-4", ModelPrice{0.80, 4.00, 0.08}},
	{"claude-3-7-sonnet", ModelPrice{3.00, 15.00, 0.30}},
	{"claude-3.7-sonnet", ModelPrice{3.00, 15.00, 0.30}},
	{"claude-3-5-sonnet", ModelPrice{3.00, 15.00, 0.30}},
	{"claude-3.5-sonnet", ModelPrice{3.00, 15.00, 0.30}},
	{"claude-3-5-haiku", ModelPrice{0.80, 4.00, 0.08}},
	{"claude-3.5-haiku", ModelPrice{0.80, 4.00, 0.08}},
	{"claude-3-opus", ModelPrice{15.00, 75.00, 1.50}},

	// Google Gemini
	{"gemini-2.5-pro", ModelPrice{1.25, 10.00, 0.31}},
	{"gemini-2.5-flash-lite", ModelPrice{0.10, 0.40, 0.025}},
	{"gemini-2.5-flash", ModelPrice{0.30, 2.50, 0.075}},
	{"gemini-2.0-flash-lite", ModelPrice{0.075, 0.30, 0.018}},
	{"gemini-2.0-flash", ModelPrice{0.10, 0.40, 0.025}},
	{"gemini-1.5-pro", ModelPrice{1.25, 5.00, 0.31}},
	{"gemini-1.5-flash", ModelPrice{0.075, 0.30, 0.018}},

	// DeepSeek
	{"deepseek-reasoner", ModelPrice{0.55, 2.19, 0.14}},
	{"deepseek-r1", ModelPrice{0.55, 2.19, 0.14}},
	{"deepseek-chat", ModelPrice{0.27, 1.10, 0.07}},
	{"deepseek-v3", ModelPrice{0.27, 1.10, 0.07}},

	// xAI Grok
	{"grok-3-mini", ModelPrice{0.30, 0.50, 0.075}},
	{"grok-3", ModelPrice{3.00, 15.00, 0.75}},
	{"grok-4", ModelPrice{3.00, 15.00, 0.75}},

	// Qwen / Llama / Mistral hosted defaults
	{"qwen3-coder", ModelPrice{0.40, 1.60, 0.10}},
	{"qwen-2.5-72b", ModelPrice{0.35, 0.40, 0.10}},
	{"qwen3", ModelPrice{0.30, 1.20, 0.08}},
	{"llama-3.3-70b", ModelPrice{0.23, 0.40, 0.06}},
	{"llama-4", ModelPrice{0.20, 0.60, 0.05}},
	{"mistral-large", ModelPrice{2.00, 6.00, 0.50}},
	{"codestral", ModelPrice{0.30, 0.90, 0.08}},
}

// PriceFor returns approximate USD/1M pricing for a model name.
func PriceFor(model string) (ModelPrice, bool) {
	m := strings.ToLower(strings.TrimSpace(model))
	if m == "" {
		return ModelPrice{}, false
	}
	if i := strings.LastIndex(m, "/"); i >= 0 {
		m = m[i+1:]
	}
	for _, r := range priceTable {
		if strings.Contains(m, r.match) {
			return r.price, true
		}
	}
	return ModelPrice{}, false
}

// EstimateCost computes approximate USD cost for the given model and token usage.
func EstimateCost(model string, u Usage) (float64, bool) {
	p, ok := PriceFor(model)
	if !ok {
		return 0, false
	}
	cached := u.CachedTokens
	if cached > u.PromptTokens {
		cached = u.PromptTokens
	}
	uncached := u.PromptTokens - cached
	usd := (float64(uncached)*p.InputPerM + float64(cached)*p.CachePerM + float64(u.CompletionTokens)*p.OutputPerM) / 1_000_000.0
	return usd, true
}

// FormatCost formats a USD cost concisely (e.g. "$0.0042", "$0.14", "$1.25").
func FormatCost(usd float64) string {
	switch {
	case usd <= 0:
		return "$0.00"
	case usd < 0.0001:
		return "<$0.0001"
	case usd < 0.01:
		return fmt.Sprintf("$%.4f", usd)
	case usd < 1:
		return fmt.Sprintf("$%.3f", usd)
	default:
		return fmt.Sprintf("$%.2f", usd)
	}
}
