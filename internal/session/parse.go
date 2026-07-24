package session

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

// Usage holds token counts from a message.
type Usage struct {
	InputTokens          int `json:"input_tokens"`
	OutputTokens         int `json:"output_tokens"`
	CacheCreationInput   int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens int `json:"cache_read_input_tokens"`
}

// ContentBlock is a single content entry inside a message.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Message is the Claude Code message envelope inside each event.
type Message struct {
	Role    string           `json:"role"`
	Model   string           `json:"model"`
	Content *json.RawMessage `json:"content"` // string or []ContentBlock
	Usage   *Usage           `json:"usage"`
}

// RawEvent is the top-level JSONL line structure.
type RawEvent struct {
	UUID             string   `json:"uuid"`
	ParentUUID       string   `json:"parentUuid"`
	Timestamp        string   `json:"timestamp"`
	IsSidechain      bool     `json:"isSidechain"`
	AttributionSkill string   `json:"attributionSkill"`
	Message          *Message `json:"message"`
	ToolUseResult    bool     `json:"toolUseResult"`
}

// NormalizedEvent is a flattened, cleaned-up representation of one transcript
// line that we actually care about (user or assistant messages with usage).
type NormalizedEvent struct {
	Role             string
	Ts               string
	UUID             string
	ParentUUID       string
	IsSidechain      bool
	AttributionSkill string
	Model            string
	IsRealUserText   bool
	TextPreview      string
	Usage            *Usage
}

// isRealUserMessage returns true when the event is a genuine user text input
// (not a tool result echo, not empty).
func isRealUserMessage(e RawEvent) bool {
	if e.Message == nil || e.Message.Role != "user" {
		return false
	}
	if e.ToolUseResult {
		return false
	}
	if e.Message.Content == nil {
		return false
	}

	return hasRealTextContent(*e.Message.Content)
}

func hasRealTextContent(raw json.RawMessage) bool {
	// Try string first
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return len(strings.TrimSpace(s)) > 0
	}

	// Try array of content blocks
	var blocks []ContentBlock
	if json.Unmarshal(raw, &blocks) != nil {
		return false
	}
	hasToolResult := false
	hasText := false
	for _, b := range blocks {
		if b.Type == "tool_result" {
			hasToolResult = true
		}
		if b.Type == "text" && len(strings.TrimSpace(b.Text)) > 0 {
			hasText = true
		}
	}
	return !hasToolResult && hasText
}

// ExtractTextPreview gets the first text content block's text, truncated.
func ExtractTextPreview(raw json.RawMessage) string {
	if raw == nil {
		return ""
	}

	var s string
	if json.Unmarshal(raw, &s) == nil {
		return strings.TrimSpace(s)
	}

	var blocks []ContentBlock
	if json.Unmarshal(raw, &blocks) != nil {
		return ""
	}
	for _, b := range blocks {
		if b.Type == "text" {
			return strings.TrimSpace(b.Text)
		}
	}
	return ""
}

// ParseSession reads a JSONL file and returns normalized events plus a
// warning count for any unparseable lines.
func ParseSession(filePath string) ([]NormalizedEvent, int) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, 0
	}
	defer f.Close()
	var events []NormalizedEvent
	warnings := 0
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var raw RawEvent

		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			warnings++
			continue
		}
		evt := normalizeEvent(raw)
		if evt != nil {
			events = append(events, *evt)
		}
	}

	if err := scanner.Err(); err != nil {
		warnings++
	}
	return events, warnings
}

func normalizeEvent(raw RawEvent) *NormalizedEvent {
	if raw.Message == nil {
		return nil
	}
	role := raw.Message.Role
	if role != "user" && role != "assistant" {
		return nil
	}

	var usage *Usage
	if raw.Message.Usage != nil {
		u := *raw.Message.Usage
		usage = &u
	}

	var textPreview string
	if raw.Message.Content != nil {
		textPreview = ExtractTextPreview(*raw.Message.Content)
	}

	return &NormalizedEvent{
		Role:             role,
		Ts:               raw.Timestamp,
		UUID:             raw.UUID,
		ParentUUID:       raw.ParentUUID,
		IsSidechain:      raw.IsSidechain,
		AttributionSkill: raw.AttributionSkill,
		Model:            raw.Message.Model,
		IsRealUserText:   isRealUserMessage(raw),
		TextPreview:      textPreview,
		Usage:            usage,
	}
}
