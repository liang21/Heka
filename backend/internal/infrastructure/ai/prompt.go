// tasks.md: T092 | spec.md: AI prompt templates
package ai

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// TestCaseGenerationPrompt is the template for generating test cases
const TestCaseGenerationPrompt = `You are an expert QA engineer. Generate comprehensive test cases based on the following requirements:

REQUIREMENTS:
{{.Requirements}}

CONTEXT:
{{.Context}}

Generate test cases in JSON format following this structure:
{
  "test_cases": [
    {
      "title": "Test case title",
      "description": "Detailed description",
      "steps": [
        {"order": 1, "action": "Action description", "expected": "Expected result"}
      ],
      "priority": "high|medium|low",
      "type": "functional|ui|api|integration",
      "tags": ["tag1", "tag2"]
    }
  ]
}

Requirements:
- Generate 5-10 diverse test cases
- Cover positive and negative scenarios
- Include edge cases and boundary conditions
- Make steps specific and actionable
- Set appropriate priorities based on business impact
- Assign relevant tags for organization`

// SmartAnalysisPrompt is the template for analyzing test results
const SmartAnalysisPrompt = `You are a test analysis expert. Analyze the following test execution results:

TEST PLAN: {{.TestPlanName}}
EXECUTION DATE: {{.ExecutionDate}}

RESULTS SUMMARY:
{{.ResultsSummary}}

DETAILED FAILURES:
{{.Failures}}

Provide a comprehensive analysis in JSON format:
{
  "analysis": {
    "overall_health": "excellent|good|fair|poor|critical",
    "pass_rate": 85.5,
    "key_insights": [
      "Insight 1",
      "Insight 2"
    ],
    "risk_areas": [
      {
        "area": "Feature name",
        "risk_level": "high|medium|low",
        "recommendation": "Specific recommendation"
      }
    ],
    "trends": {
      "improving": ["Feature A"],
      "regressing": ["Feature B"],
      "stable": ["Feature C"]
    },
    "suggested_actions": [
      "Action 1",
      "Action 2"
    ]
  }
}

Focus on:
- Identifying patterns in failures
- Highlighting high-risk areas
- Detecting regression or improvement trends
- Providing actionable recommendations
- Considering business impact`

// RAGQueryPrompt is the template for RAG-based queries
const RAGQueryPrompt = `You are a knowledgeable test engineer assistant. Answer the following question based on the provided context:

QUESTION: {{.Question}}

RELEVANT CONTEXT:
{{.Context}}

Provide a helpful, accurate response. If the context doesn't contain enough information to answer the question, say so explicitly.

Cite specific parts of the context when relevant.`

// CodeReviewPrompt is the template for AI code review
const CodeReviewPrompt = `You are a senior Go code reviewer. Review the following code:

FILE: {{.FilePath}}
LANGUAGE: Go

CODE:
{{.Code}}

Provide a code review in JSON format:
{
  "review": {
    "overall_score": 1-10,
    "summary": "Brief summary",
    "strengths": ["Strength 1", "Strength 2"],
    "issues": [
      {
        "severity": "critical|major|minor",
        "category": "security|performance|maintainability|bug|style",
        "line": 123,
        "description": "Issue description",
        "suggestion": "How to fix"
      }
    ],
    "suggestions": [
      "General suggestion 1",
      "General suggestion 2"
    ]
  }
}

Focus on:
- Security vulnerabilities
- Performance issues
- Code maintainability
- Potential bugs
- Go best practices
- Concurrency safety`

// BuildPrompt replaces variables in a template with provided values
func BuildPrompt(template string, vars map[string]string) string {
	result := template

	// Replace all variables in the format {{.VariableName}}
	re := regexp.MustCompile(`\{\{\.(\w+)\}\}`)

	matches := re.FindAllStringSubmatch(result, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		varName := match[1]
		varValue, exists := vars[varName]
		if !exists {
			varValue = fmt.Sprintf("[MISSING: %s]", varName)
		}

		result = strings.ReplaceAll(result, match[0], varValue)
	}

	return result
}

// BuildTestCasePrompt builds a test case generation prompt
func BuildTestCasePrompt(requirements, context string) string {
	return BuildPrompt(TestCaseGenerationPrompt, map[string]string{
		"Requirements": requirements,
		"Context":      context,
	})
}

// BuildAnalysisPrompt builds a test analysis prompt
func BuildAnalysisPrompt(testPlanName, executionDate, resultsSummary, failures string) string {
	return BuildPrompt(SmartAnalysisPrompt, map[string]string{
		"TestPlanName":   testPlanName,
		"ExecutionDate":  executionDate,
		"ResultsSummary": resultsSummary,
		"Failures":       failures,
	})
}

// BuildRAGQueryPrompt builds a RAG query prompt
func BuildRAGQueryPrompt(question, context string) string {
	return BuildPrompt(RAGQueryPrompt, map[string]string{
		"Question": question,
		"Context":  context,
	})
}

// BuildCodeReviewPrompt builds a code review prompt
func BuildCodeReviewPrompt(filePath, code string) string {
	return BuildPrompt(CodeReviewPrompt, map[string]string{
		"FilePath": filePath,
		"Code":     code,
	})
}

// ValidatePrompt checks if a prompt template is valid
func ValidatePrompt(template string) error {
	// Check for unbalanced braces
	openBraces := strings.Count(template, "{{")
	closeBraces := strings.Count(template, "}}")

	if openBraces != closeBraces {
		return fmt.Errorf("unbalanced braces in template: %d open, %d close", openBraces, closeBraces)
	}

	return nil
}

// ExtractVariables extracts variable names from a template
func ExtractVariables(template string) []string {
	re := regexp.MustCompile(`\{\{\.(\w+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	vars := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) >= 2 {
			varName := match[1]
			if !seen[varName] {
				seen[varName] = true
				vars = append(vars, varName)
			}
		}
	}

	return vars
}

// FormatPrompt formats a prompt for display (for debugging)
func FormatPrompt(template string, vars map[string]string) string {
	var buf bytes.Buffer

	buf.WriteString("=== PROMPT TEMPLATE ===\n")
	buf.WriteString(template)
	buf.WriteString("\n\n=== VARIABLES ===\n")
	for key, value := range vars {
		buf.WriteString(fmt.Sprintf("%s = %s\n", key, value))
	}
	buf.WriteString("\n=== RENDERED PROMPT ===\n")
	buf.WriteString(BuildPrompt(template, vars))

	return buf.String()
}
