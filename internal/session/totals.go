package session

import (
	"slices"
	"sort"
	"strings"
)

// DefaultModelLimits maps model family keywords to their context window size.
var DefaultModelLimits = map[string]int{
	"sonnet": 200000,
	"opus":   200000,
	"haiku":  200000,
}

// SkillTotal holds one row of the by-skill breakdown.
type SkillTotal struct {
	Skill string
	Total int
}

// ContextWindow holds the calculated context window state.
type ContextWindow struct {
	TotalInputTokens int
	WindowSize       int
	UsedPercentage   float64 // rounded to one decimal place
}

// Totals is the fully computed aggregation from a session.
type Totals struct {
	Usage         Usage
	BySkill       []SkillTotal   // sorted descending
	TopTurns      []Turn         // sorted descending by total tokens
	ContextWindow *ContextWindow // nil when no assistant events exist yet
}

// ComputeOpts allows overriding defaults for ComputeTotals.
type ComputeOpts struct {
	ModelLimits map[string]int
	TopN        int // not used internally but reserved
}

// ComputeTotals aggregates token usage across all turns: grand totals,
// per-skill breakdown, top prompt ranking, and context window calculation.
func ComputeTotals(events []NormalizedEvent, turns []Turn, opts ...func(*ComputeOpts)) Totals {
	var cfg ComputeOpts

	for _, fn := range opts {
		fn(&cfg)
	}

	modelLimits := DefaultModelLimits

	if cfg.ModelLimits != nil {
		modelLimits = cfg.ModelLimits
	}

	var totals Usage
	bySkillMap := make(map[string]int)

	for _, turn := range turns {
		totals.InputTokens += turn.Usage.InputTokens
		totals.OutputTokens += turn.Usage.OutputTokens
		totals.CacheCreationInput += turn.Usage.CacheCreationInput
		totals.CacheReadInputTokens += turn.Usage.CacheReadInputTokens

		for skill, tokens := range turn.BySkill {
			bySkillMap[skill] += tokens
		}
	}

	var bySkill []SkillTotal

	for skill, total := range bySkillMap {
		bySkill = append(bySkill, SkillTotal{Skill: skill, Total: total})
	}

	sort.Slice(bySkill, func(i, j int) bool {
		return bySkill[i].Total > bySkill[j].Total
	})

	topTurns := make([]Turn, len(turns))
	copy(topTurns, turns)

	sort.Slice(topTurns, func(i, j int) bool {
		totalI := topTurns[i].Usage.InputTokens + topTurns[i].Usage.OutputTokens +
			topTurns[i].Usage.CacheCreationInput + topTurns[i].Usage.CacheReadInputTokens
		totalJ := topTurns[j].Usage.InputTokens + topTurns[j].Usage.OutputTokens +
			topTurns[j].Usage.CacheCreationInput + topTurns[j].Usage.CacheReadInputTokens

		return totalI > totalJ
	})

	// Context window: find the last (most recent) main assistant event.
	var ctx *ContextWindow

	for _, e := range slices.Backward(events) {
		if e.Role == "assistant" && !e.IsSidechain && e.Usage != nil && e.Model != "" {
			totalInput := e.Usage.InputTokens + e.Usage.CacheReadInputTokens + e.Usage.CacheCreationInput
			limit := resolveModelLimit(e.Model, modelLimits)
			pct := float64(totalInput) / float64(limit) * 100
			// Round to one decimal
			pct = float64(int(pct*10)) / 10
			ctx = &ContextWindow{
				TotalInputTokens: totalInput,
				WindowSize:       limit,
				UsedPercentage:   pct,
			}
			break
		}
	}

	return Totals{
		Usage:         totals,
		BySkill:       bySkill,
		TopTurns:      topTurns,
		ContextWindow: ctx,
	}
}

func resolveModelLimit(modelID string, limits map[string]int) int {
	if modelID == "" {
		return 200000
	}

	lower := strings.ToLower(modelID)

	for key, limit := range limits {
		if strings.Contains(lower, strings.ToLower(key)) {
			return limit
		}
	}

	return 200000
}
