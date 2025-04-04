package main

import (
	"encoding/json"
	"strings"
)

// Removes unncessery post/pre fixes
func extractJSONfromLLM(text string, v any) error {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimSuffix(text, "```")
	return json.Unmarshal([]byte(text), v)
}
