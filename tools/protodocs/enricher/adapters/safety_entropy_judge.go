package adapters

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/kyivinua/root/tools/protodocs/enricher"
)

// EntropyJudgeSafetyGuard implements SafetyGuard using semantic entropy and LLM-as-judge
type EntropyJudgeSafetyGuard struct {
	llm                LLMClient
	judgeModel         string
	entropyThreshold   float64
	confidenceThreshold float64
	piiChecks          bool
	pciChecks          bool
}

// NewEntropyJudgeSafetyGuard creates a new entropy+judge safety guard
func NewEntropyJudgeSafetyGuard(
	llm LLMClient,
	judgeModel string,
	entropyThreshold float64,
	confidenceThreshold float64,
	piiChecks bool,
	pciChecks bool,
) *EntropyJudgeSafetyGuard {
	return &EntropyJudgeSafetyGuard{
		llm:                 llm,
		judgeModel:          judgeModel,
		entropyThreshold:    entropyThreshold,
		confidenceThreshold: confidenceThreshold,
		piiChecks:           piiChecks,
		pciChecks:           pciChecks,
	}
}

// Validate validates enriched content for safety
func (sg *EntropyJudgeSafetyGuard) Validate(
	ctx context.Context,
	original string,
	enriched string,
	target enricher.EnrichmentTarget,
) (*enricher.SafetyReport, error) {
	report := &enricher.SafetyReport{
		Passed:         true,
		Confidence:     1.0,
		FailureReasons: []string{},
		DetectedIssues: []enricher.SafetyIssue{},
		Timestamp:      time.Now(),
	}

	// Step 1: PII/PCI checks
	if sg.piiChecks {
		if issues := sg.detectPII(enriched); len(issues) > 0 {
			report.Passed = false
			report.DetectedIssues = append(report.DetectedIssues, issues...)
			report.FailureReasons = append(report.FailureReasons, "PII detected in enriched content")
		}
	}

	if sg.pciChecks {
		if issues := sg.detectPCI(enriched); len(issues) > 0 {
			report.Passed = false
			report.DetectedIssues = append(report.DetectedIssues, issues...)
			report.FailureReasons = append(report.FailureReasons, "PCI-DSS sensitive data detected")
		}
	}

	// Step 2: Semantic entropy check
	entropyScore := sg.calculateSemanticEntropy(enriched)
	report.EntropyScore = entropyScore

	if entropyScore > sg.entropyThreshold {
		report.Passed = false
		report.FailureReasons = append(report.FailureReasons,
			fmt.Sprintf("high semantic entropy: %.2f > %.2f", entropyScore, sg.entropyThreshold))
		report.Confidence = 1.0 - (entropyScore - sg.entropyThreshold)
	}

	// Step 3: LLM-as-judge validation
	if sg.llm != nil {
		judgeVerdict, judgeConfidence, err := sg.runLLMJudge(ctx, original, enriched, target)
		if err != nil {
			return report, fmt.Errorf("LLM judge: %w", err)
		}

		report.JudgeVerdict = judgeVerdict
		report.Confidence = math.Min(report.Confidence, judgeConfidence)

		if !strings.Contains(strings.ToLower(judgeVerdict), "faithful") {
			report.Passed = false
			report.FailureReasons = append(report.FailureReasons,
				fmt.Sprintf("LLM judge rejected: %s", judgeVerdict))
		}

		if judgeConfidence < sg.confidenceThreshold {
			report.Passed = false
			report.FailureReasons = append(report.FailureReasons,
				fmt.Sprintf("low confidence: %.2f < %.2f", judgeConfidence, sg.confidenceThreshold))
		}
	}

	return report, nil
}

// detectPII detects personally identifiable information
func (sg *EntropyJudgeSafetyGuard) detectPII(content string) []enricher.SafetyIssue {
	issues := []enricher.SafetyIssue{}

	// Email pattern
	emailRegex := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
	if matches := emailRegex.FindAllString(content, -1); len(matches) > 0 {
		issues = append(issues, enricher.SafetyIssue{
			Type:     "PII",
			Severity: "HIGH",
			Message:  fmt.Sprintf("Email addresses detected: %v", matches),
			Location: "content",
		})
	}

	// Phone number pattern (US format)
	phoneRegex := regexp.MustCompile(`\b(\+1[-.\s]?)?(\(?\d{3}\)?[-.\s]?)?\d{3}[-.\s]?\d{4}\b`)
	if matches := phoneRegex.FindAllString(content, -1); len(matches) > 0 {
		issues = append(issues, enricher.SafetyIssue{
			Type:     "PII",
			Severity: "HIGH",
			Message:  fmt.Sprintf("Phone numbers detected: %v", matches),
			Location: "content",
		})
	}

	// SSN pattern (US)
	ssnRegex := regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	if matches := ssnRegex.FindAllString(content, -1); len(matches) > 0 {
		issues = append(issues, enricher.SafetyIssue{
			Type:     "PII",
			Severity: "CRITICAL",
			Message:  "SSN-like patterns detected",
			Location: "content",
		})
	}

	// IP address pattern
	ipRegex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	if matches := ipRegex.FindAllString(content, -1); len(matches) > 0 {
		// Filter out common non-PII IPs like 127.0.0.1, 0.0.0.0
		filteredMatches := []string{}
		for _, ip := range matches {
			if ip != "127.0.0.1" && ip != "0.0.0.0" && !strings.HasPrefix(ip, "10.") {
				filteredMatches = append(filteredMatches, ip)
			}
		}
		if len(filteredMatches) > 0 {
			issues = append(issues, enricher.SafetyIssue{
				Type:     "PII",
				Severity: "MEDIUM",
				Message:  fmt.Sprintf("IP addresses detected: %v", filteredMatches),
				Location: "content",
			})
		}
	}

	return issues
}

// detectPCI detects PCI-DSS sensitive data
func (sg *EntropyJudgeSafetyGuard) detectPCI(content string) []enricher.SafetyIssue {
	issues := []enricher.SafetyIssue{}

	// Credit card number pattern (Luhn algorithm not validated here for simplicity)
	ccRegex := regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`)
	if matches := ccRegex.FindAllString(content, -1); len(matches) > 0 {
		issues = append(issues, enricher.SafetyIssue{
			Type:     "PCI-DSS",
			Severity: "CRITICAL",
			Message:  "Credit card-like patterns detected",
			Location: "content",
		})
	}

	// CVV pattern
	cvvRegex := regexp.MustCompile(`\b\d{3,4}\b`)
	// This is too broad, so we only flag if found in context with card-related keywords
	if strings.Contains(strings.ToLower(content), "cvv") || strings.Contains(strings.ToLower(content), "cvc") {
		if matches := cvvRegex.FindAllString(content, -1); len(matches) > 0 {
			issues = append(issues, enricher.SafetyIssue{
				Type:     "PCI-DSS",
				Severity: "CRITICAL",
				Message:  "CVV/CVC patterns detected",
				Location: "content",
			})
		}
	}

	return issues
}

// calculateSemanticEntropy calculates semantic entropy (von Neumann entropy approximation)
func (sg *EntropyJudgeSafetyGuard) calculateSemanticEntropy(content string) float64 {
	// Simplified semantic entropy based on word distribution
	// Real implementation would use embedding-based clustering

	words := strings.Fields(strings.ToLower(content))
	if len(words) == 0 {
		return 0.0
	}

	// Count word frequencies
	wordFreq := make(map[string]int)
	for _, word := range words {
		wordFreq[word]++
	}

	// Calculate Shannon entropy as proxy for semantic entropy
	entropy := 0.0
	totalWords := float64(len(words))

	for _, freq := range wordFreq {
		probability := float64(freq) / totalWords
		if probability > 0 {
			entropy -= probability * math.Log2(probability)
		}
	}

	// Normalize to [0, 1]
	maxEntropy := math.Log2(float64(len(wordFreq)))
	if maxEntropy > 0 {
		entropy = entropy / maxEntropy
	}

	return entropy
}

// runLLMJudge runs LLM-as-judge to validate faithfulness
func (sg *EntropyJudgeSafetyGuard) runLLMJudge(
	ctx context.Context,
	original string,
	enriched string,
	target enricher.EnrichmentTarget,
) (verdict string, confidence float64, err error) {
	// Build judge prompt
	judgePrompt := fmt.Sprintf(`You are a faithful documentation reviewer. Evaluate if the enriched documentation is faithful to the original and adds value without hallucination.

<original_documentation>
%s
</original_documentation>

<enriched_documentation>
%s
</enriched_documentation>

<target_context>
Type: %s
Identifier: %s
</target_context>

Evaluate the enriched documentation and respond in this exact format:

VERDICT: [FAITHFUL|UNFAITHFUL]
CONFIDENCE: [0.0-1.0]
REASONING: [Brief explanation]

Criteria:
1. Does enriched content contradict original?
2. Does enriched content introduce unverifiable claims?
3. Does enriched content maintain technical accuracy?
4. Does enriched content add meaningful value?`, original, enriched, fmt.Sprintf("%T", target), target.GetIdentifier())

	// Call LLM judge
	config := map[string]interface{}{
		"temperature": 0.0,
		"max_tokens":  500,
	}

	response, err := sg.llm.GenerateCompletionWithConfig(ctx, judgePrompt, config)
	if err != nil {
		return "", 0.0, fmt.Errorf("judge LLM call: %w", err)
	}

	// Parse response
	verdict = "UNFAITHFUL" // Default to conservative
	confidence = 0.0

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VERDICT:") {
			verdict = strings.TrimSpace(strings.TrimPrefix(line, "VERDICT:"))
		} else if strings.HasPrefix(line, "CONFIDENCE:") {
			confStr := strings.TrimSpace(strings.TrimPrefix(line, "CONFIDENCE:"))
			fmt.Sscanf(confStr, "%f", &confidence)
		}
	}

	return verdict, confidence, nil
}

// NoOpSafetyGuard is a no-op implementation for when safety checks are disabled
type NoOpSafetyGuard struct{}

// NewNoOpSafetyGuard creates a no-op safety guard
func NewNoOpSafetyGuard() *NoOpSafetyGuard {
	return &NoOpSafetyGuard{}
}

// Validate always passes
func (sg *NoOpSafetyGuard) Validate(
	ctx context.Context,
	original string,
	enriched string,
	target enricher.EnrichmentTarget,
) (*enricher.SafetyReport, error) {
	return &enricher.SafetyReport{
		Passed:         true,
		Confidence:     1.0,
		FailureReasons: []string{},
		DetectedIssues: []enricher.SafetyIssue{},
		Timestamp:      time.Now(),
	}, nil
}
