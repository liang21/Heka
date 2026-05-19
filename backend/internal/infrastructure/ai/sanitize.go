// tasks.md: T093 | spec.md: Input sanitization and output validation
package ai

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/liang21/heka/internal/domain/shared"
)

// Prompt injection patterns to detect and remove
var injectionPatterns = []string{
	"(?i)ignore (previous|all|above) (instructions?|commands?|directives?)",
	"(?i)disregard (previous|all|above) (instructions?|commands?|directives?)",
	"(?i)system[:\\s]*override",
	"(?i)\\bclass\\b.*\\bconstructor\\b", // JavaScript injection attempt
	"(?i)<script[^>]*>.*</script>",       // Script injection
	"(?i)javascript:",
	"(?i)on\\w+\\s*=",                        // Event handler injection
	"(?i)```\\s*(json|javascript|js|python)", // Code fence abuse
	"(?i)<<\\s*\\w+",                         // Heredoc injection
	"(?i)eval\\s*\\(",
	"(?i)exec\\s*\\(",
}

// Malicious patterns that might cause issues
var maliciousPatterns = []string{
	"(?i)\\bdrop\\s+table\\b",
	"(?i)\\bdelete\\s+from\\b",
	"(?i)\\btruncate\\b",
	"(?i)\\bunion\\s+select\\b",
	"(?i)\\bxp_cmdshell\\b",
	"(?i)\\bsp_executesql\\b",
}

// SanitizeInput removes potentially dangerous content from user input
func SanitizeInput(input string) string {
	if input == "" {
		return input
	}

	result := input

	// Remove prompt injection patterns
	for _, pattern := range injectionPatterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, "[REDACTED]")
	}

	// Remove malicious SQL patterns
	for _, pattern := range maliciousPatterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, "[REDACTED]")
	}

	// Limit length to prevent token abuse
	maxLength := 50000 // 50k characters
	if len(result) > maxLength {
		result = result[:maxLength] + "... [TRUNCATED]"
	}

	// Remove excessive whitespace
	result = compressWhitespace(result)

	// Remove invisible control characters (except newline, tab, carriage return)
	result = removeControlCharacters(result)

	return strings.TrimSpace(result)
}

// SanitizeMessageList sanitizes a list of messages
func SanitizeMessageList(messages []Message) []Message {
	if messages == nil {
		return nil
	}

	sanitized := make([]Message, len(messages))
	for i, msg := range messages {
		sanitized[i] = Message{
			Role:    msg.Role,
			Content: SanitizeInput(msg.Content),
		}
	}

	return sanitized
}

// ValidateAIOutput validates AI-generated output for safety and correctness
func ValidateAIOutput(output string) error {
	if output == "" {
		return shared.ErrAIInvalidInput
	}

	// Check for obvious prompt injection in output
	for _, pattern := range injectionPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(output) {
			return fmt.Errorf("%w: output contains potentially harmful content", shared.ErrAIInvalidInput)
		}
	}

	// Check for code injection attempts
	if containsSuspiciousCode(output) {
		return fmt.Errorf("%w: output contains suspicious code patterns", shared.ErrAIInvalidInput)
	}

	return nil
}

// ValidateJSONOutput validates and parses JSON output
func ValidateJSONOutput(output string) (map[string]interface{}, error) {
	output = strings.TrimSpace(output)

	// Try to extract JSON if embedded in markdown code fences
	if strings.HasPrefix(output, "```") {
		re := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			output = strings.TrimSpace(matches[1])
		}
	}

	// Validate JSON structure
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("%w: invalid JSON: %v", shared.ErrAIInvalidInput, err)
	}

	// Check for required top-level keys
	if len(result) == 0 {
		return nil, fmt.Errorf("%w: empty JSON object", shared.ErrAIInvalidInput)
	}

	return result, nil
}

// ValidateAndParseOutput validates output and parses into target structure
func ValidateAndParseOutput(output string, target interface{}) error {
	output = strings.TrimSpace(output)

	// Try to extract JSON if embedded in markdown code fences
	if strings.HasPrefix(output, "```") {
		re := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			output = strings.TrimSpace(matches[1])
		}
	}

	// Parse JSON
	if err := json.Unmarshal([]byte(output), target); err != nil {
		return fmt.Errorf("%w: failed to parse output: %v", shared.ErrAIInvalidInput, err)
	}

	return nil
}

// DetectPromptInjection checks if input contains prompt injection attempts
func DetectPromptInjection(input string) bool {
	for _, pattern := range injectionPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(input) {
			return true
		}
	}
	return false
}

// DetectSQLInjection checks if input contains SQL injection attempts
func DetectSQLInjection(input string) bool {
	for _, pattern := range maliciousPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(input) {
			return true
		}
	}
	return false
}

// compressWhitespace reduces multiple consecutive whitespace to single space
func compressWhitespace(s string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, " ")
}

// removeControlCharacters removes invisible control characters
func removeControlCharacters(s string) string {
	var result []rune
	for _, r := range s {
		if r == '\n' || r == '\t' || r == '\r' {
			result = append(result, r)
		} else if !unicode.IsControl(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// containsSuspiciousCode checks for potentially dangerous code patterns
func containsSuspiciousCode(s string) bool {
	suspicious := []string{
		"<script",
		"javascript:",
		"eval(",
		"exec(",
		"system(",
		"shell_exec",
		"passthru",
		"proc_open",
		"popen(",
	}

	lower := strings.ToLower(s)
	for _, pattern := range suspicious {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// TruncateContent truncates content to a maximum length at word boundaries
func TruncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}

	// Truncate at word boundary
	truncated := content[:maxLen]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

// EscapeMarkdown escapes markdown special characters for safe display
func EscapeMarkdown(text string) string {
	chars := []string{"\\", "`", "*", "_", "{", "}", "[", "]", "(", ")", "#", "+", "-", ".", "!"}

	result := text
	for _, char := range chars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}

	return result
}
