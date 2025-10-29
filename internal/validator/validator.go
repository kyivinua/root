// Package validator provides quality validation for documentation.
package validator

import (
	"fmt"
	"strings"

	"github.com/kyivinua/docgen-tool/internal/docgen"
)

// Validator validates documentation quality.
type Validator struct {
	config Config
}

// Config represents validator configuration.
type Config struct {
	CoverageTarget           float64
	DescriptionQualityTarget float64
	MinMethodCoverage        float64
	MinFieldCoverage         float64
	AutoFix                  bool
	FailOnQualityGate        bool
}

// NewValidator creates a new validator.
func NewValidator(config Config) *Validator {
	return &Validator{config: config}
}

// Validate validates a service documentation.
func (v *Validator) Validate(service *docgen.Service) (*docgen.ServiceQuality, []docgen.QualityIssue, error) {
	var issues []docgen.QualityIssue

	// Calculate method coverage
	methodCoverage := v.calculateMethodCoverage(service)
	if methodCoverage < v.config.MinMethodCoverage {
		issues = append(issues, docgen.QualityIssue{
			Severity:    "warning",
			Type:        "coverage",
			Message:     fmt.Sprintf("Method coverage %.1f%% is below target %.1f%%", methodCoverage, v.config.MinMethodCoverage),
			Location:    service.Name,
			Suggestion:  "Add descriptions to undocumented methods",
			AutoFixable: v.config.AutoFix,
		})
	}

	// Calculate field coverage
	fieldCoverage := v.calculateFieldCoverage(service)
	if fieldCoverage < v.config.MinFieldCoverage {
		issues = append(issues, docgen.QualityIssue{
			Severity:    "warning",
			Type:        "coverage",
			Message:     fmt.Sprintf("Field coverage %.1f%% is below target %.1f%%", fieldCoverage, v.config.MinFieldCoverage),
			Location:    service.Name,
			Suggestion:  "Add descriptions to undocumented fields",
			AutoFixable: v.config.AutoFix,
		})
	}

	// Calculate overall coverage
	coverage := (methodCoverage + fieldCoverage) / 2

	// Calculate description quality
	descQuality := v.calculateDescriptionQuality(service)
	if descQuality < v.config.DescriptionQualityTarget {
		issues = append(issues, docgen.QualityIssue{
			Severity:    "info",
			Type:        "quality",
			Message:     fmt.Sprintf("Description quality %.1f%% is below target %.1f%%", descQuality, v.config.DescriptionQualityTarget),
			Location:    service.Name,
			Suggestion:  "Improve description quality by adding more details",
			AutoFixable: v.config.AutoFix,
		})
	}

	// Count documented items
	documentedMethods := 0
	for _, method := range service.Methods {
		if method.Description != "" {
			documentedMethods++
		}
	}

	totalFields := 0
	documentedFields := 0
	for _, message := range service.Messages {
		for _, field := range message.Fields {
			totalFields++
			if field.Description != "" {
				documentedFields++
			}
		}
	}

	quality := &docgen.ServiceQuality{
		ServiceName:        service.Name,
		Coverage:           coverage,
		DescriptionQuality: descQuality,
		MethodCount:        len(service.Methods),
		DocumentedMethods:  documentedMethods,
		FieldCount:         totalFields,
		DocumentedFields:   documentedFields,
	}

	return quality, issues, nil
}

// ValidateAll validates all services and generates a quality report.
func (v *Validator) ValidateAll(services []docgen.Service) (*docgen.QualityReport, error) {
	report := &docgen.QualityReport{
		ServiceMetrics: make(map[string]docgen.ServiceQuality),
		Issues:         []docgen.QualityIssue{},
		Passed:         true,
	}

	var totalCoverage, totalQuality float64

	for i := range services {
		quality, issues, err := v.Validate(&services[i])
		if err != nil {
			return nil, fmt.Errorf("failed to validate service %s: %w", services[i].Name, err)
		}

		report.ServiceMetrics[services[i].Name] = *quality
		report.Issues = append(report.Issues, issues...)

		totalCoverage += quality.Coverage
		totalQuality += quality.DescriptionQuality
	}

	// Calculate averages
	if len(services) > 0 {
		report.CoverageScore = totalCoverage / float64(len(services))
		report.DescriptionQuality = totalQuality / float64(len(services))

		// Calculate method and field coverage
		totalMethods := 0
		totalDocumentedMethods := 0
		totalFields := 0
		totalDocumentedFields := 0

		for _, q := range report.ServiceMetrics {
			totalMethods += q.MethodCount
			totalDocumentedMethods += q.DocumentedMethods
			totalFields += q.FieldCount
			totalDocumentedFields += q.DocumentedFields
		}

		if totalMethods > 0 {
			report.MethodCoverage = float64(totalDocumentedMethods) / float64(totalMethods) * 100
		}
		if totalFields > 0 {
			report.FieldCoverage = float64(totalDocumentedFields) / float64(totalFields) * 100
		}
	}

	// Check if passed quality gates
	if report.CoverageScore < v.config.CoverageTarget {
		report.Passed = false
		if v.config.FailOnQualityGate {
			return report, fmt.Errorf("quality gate failed: coverage %.1f%% is below target %.1f%%",
				report.CoverageScore, v.config.CoverageTarget)
		}
	}

	if report.DescriptionQuality < v.config.DescriptionQualityTarget {
		report.Passed = false
		if v.config.FailOnQualityGate {
			return report, fmt.Errorf("quality gate failed: description quality %.1f%% is below target %.1f%%",
				report.DescriptionQuality, v.config.DescriptionQualityTarget)
		}
	}

	return report, nil
}

// calculateMethodCoverage calculates the percentage of documented methods.
func (v *Validator) calculateMethodCoverage(service *docgen.Service) float64 {
	if len(service.Methods) == 0 {
		return 100.0
	}

	documented := 0
	for _, method := range service.Methods {
		if method.Description != "" {
			documented++
		}
	}

	return float64(documented) / float64(len(service.Methods)) * 100
}

// calculateFieldCoverage calculates the percentage of documented fields.
func (v *Validator) calculateFieldCoverage(service *docgen.Service) float64 {
	totalFields := 0
	documentedFields := 0

	for _, message := range service.Messages {
		for _, field := range message.Fields {
			totalFields++
			if field.Description != "" {
				documentedFields++
			}
		}
	}

	if totalFields == 0 {
		return 100.0
	}

	return float64(documentedFields) / float64(totalFields) * 100
}

// calculateDescriptionQuality calculates the quality score of descriptions.
func (v *Validator) calculateDescriptionQuality(service *docgen.Service) float64 {
	var scores []float64

	// Service description
	if service.Description != "" {
		scores = append(scores, scoreDescription(service.Description))
	}

	// Method descriptions
	for _, method := range service.Methods {
		if method.Description != "" {
			scores = append(scores, scoreDescription(method.Description))
		}
	}

	// Message descriptions
	for _, message := range service.Messages {
		if message.Description != "" {
			scores = append(scores, scoreDescription(message.Description))
		}
		for _, field := range message.Fields {
			if field.Description != "" {
				scores = append(scores, scoreDescription(field.Description))
			}
		}
	}

	if len(scores) == 0 {
		return 0.0
	}

	// Calculate average
	var total float64
	for _, score := range scores {
		total += score
	}

	return total / float64(len(scores))
}

// scoreDescription scores the quality of a description.
func scoreDescription(desc string) float64 {
	if desc == "" {
		return 0.0
	}

	score := 50.0 // Base score for having a description

	// Length bonus (up to 20 points)
	words := len(strings.Fields(desc))
	if words >= 10 {
		score += 20.0
	} else if words >= 5 {
		score += 10.0
	}

	// Sentence structure (up to 15 points)
	sentences := strings.Count(desc, ".") + strings.Count(desc, "!") + strings.Count(desc, "?")
	if sentences >= 2 {
		score += 15.0
	} else if sentences >= 1 {
		score += 7.5
	}

	// Capitalization (up to 5 points)
	if len(desc) > 0 && desc[0] >= 'A' && desc[0] <= 'Z' {
		score += 5.0
	}

	// Examples or details (up to 10 points)
	if strings.Contains(strings.ToLower(desc), "example") ||
		strings.Contains(strings.ToLower(desc), "e.g.") ||
		strings.Contains(desc, ":") {
		score += 10.0
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// AutoFix attempts to automatically fix quality issues.
func (v *Validator) AutoFix(service *docgen.Service, issues []docgen.QualityIssue) error {
	if !v.config.AutoFix {
		return nil
	}

	for _, issue := range issues {
		if !issue.AutoFixable {
			continue
		}

		// Add placeholder descriptions
		switch issue.Type {
		case "coverage":
			if strings.Contains(issue.Message, "Method") {
				v.fixMethodDescriptions(service)
			} else if strings.Contains(issue.Message, "Field") {
				v.fixFieldDescriptions(service)
			}
		}
	}

	return nil
}

// fixMethodDescriptions adds placeholder descriptions to methods.
func (v *Validator) fixMethodDescriptions(service *docgen.Service) {
	for i := range service.Methods {
		if service.Methods[i].Description == "" {
			service.Methods[i].Description = fmt.Sprintf("%s method for %s service.", service.Methods[i].Name, service.Name)
		}
	}
}

// fixFieldDescriptions adds placeholder descriptions to fields.
func (v *Validator) fixFieldDescriptions(service *docgen.Service) {
	for i := range service.Messages {
		for j := range service.Messages[i].Fields {
			if service.Messages[i].Fields[j].Description == "" {
				service.Messages[i].Fields[j].Description = fmt.Sprintf("%s field.", service.Messages[i].Fields[j].Name)
			}
		}
	}
}
