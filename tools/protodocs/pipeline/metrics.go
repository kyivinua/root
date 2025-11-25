package pipeline

import (
	"time"
)

// DiscoveryMetrics contains metrics and statistics about discovery process
type DiscoveryMetrics struct {
	// Discovery phase
	FilesFound         int           `json:"files_found"`
	FilesExcluded      int           `json:"files_excluded"`
	FilesParsed        int           `json:"files_parsed"`
	ParseErrors        int           `json:"parse_errors"`

	// Grouping phase
	ServicesFound      int           `json:"services_found"`
	TotalServiceDefs   int           `json:"total_service_defs"`

	// Performance
	Duration           time.Duration `json:"duration_ms"`
	ParseDuration      time.Duration `json:"parse_duration_ms"`
	GroupingDuration   time.Duration `json:"grouping_duration_ms"`

	// By detection strategy
	DetectedByDir      int           `json:"detected_by_directory"`
	DetectedByService  int           `json:"detected_by_service"`
	DetectedByPackage  int           `json:"detected_by_package"`

	// Errors
	Errors             []string      `json:"errors,omitempty"`
	Warnings           []string      `json:"warnings,omitempty"`
}

// ConsolidationMetrics contains metrics about consolidation process
type ConsolidationMetrics struct {
	ServicesProcessed  int           `json:"services_processed"`
	TotalFilesCopied   int           `json:"total_files_copied"`
	TotalErrors        int           `json:"total_errors"`
	BufConfigsCreated  int           `json:"buf_configs_created"`
	ReadmesCreated     int           `json:"readmes_created"`
	Duration           time.Duration `json:"duration_ms"`

	// Per-service metrics
	ServiceMetrics     map[string]*ServiceConsolidationMetrics `json:"service_metrics,omitempty"`
}

// ServiceConsolidationMetrics contains per-service consolidation metrics
type ServiceConsolidationMetrics struct {
	ServiceName    string        `json:"service_name"`
	FilesCopied    int           `json:"files_copied"`
	Errors         int           `json:"errors"`
	Duration       time.Duration `json:"duration_ms"`
	OutputSize     int64         `json:"output_size_bytes"`
}

// ProgressCallback is called during discovery to report progress
type ProgressCallback func(phase string, current, total int, message string)

// Logger interface for discovery operations
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// DefaultLogger provides a simple stdout logger
type DefaultLogger struct{}

func (l *DefaultLogger) Debug(msg string, args ...interface{}) {}
func (l *DefaultLogger) Info(msg string, args ...interface{})  {}
func (l *DefaultLogger) Warn(msg string, args ...interface{})  {}
func (l *DefaultLogger) Error(msg string, args ...interface{}) {}

// NewMetrics creates a new metrics instance
func NewMetrics() *DiscoveryMetrics {
	return &DiscoveryMetrics{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}
}

// AddError adds an error to metrics
func (m *DiscoveryMetrics) AddError(err error) {
	if err != nil {
		m.ParseErrors++
		m.Errors = append(m.Errors, err.Error())
	}
}

// AddWarning adds a warning to metrics
func (m *DiscoveryMetrics) AddWarning(msg string) {
	m.Warnings = append(m.Warnings, msg)
}

// Summary returns a human-readable summary
func (m *DiscoveryMetrics) Summary() string {
	return formatMetricsSummary(
		"Discovery completed",
		"Files", m.FilesFound,
		"Excluded", m.FilesExcluded,
		"Parsed", m.FilesParsed,
		"Services", m.ServicesFound,
		"Errors", m.ParseErrors,
		"Duration", m.Duration,
	)
}

func formatMetricsSummary(title string, args ...interface{}) string {
	result := title + ": "
	for i := 0; i < len(args); i += 2 {
		if i > 0 {
			result += ", "
		}
		result += formatArg(args[i], args[i+1])
	}
	return result
}

func formatArg(key, value interface{}) string {
	switch v := value.(type) {
	case time.Duration:
		return formatDuration(key.(string), v)
	case int:
		return formatInt(key.(string), v)
	default:
		return formatDefault(key.(string), v)
	}
}

func formatDuration(key string, d time.Duration) string {
	return key + "=" + d.String()
}

func formatInt(key string, i int) string {
	return key + "=" + intToString(i)
}

func formatDefault(key string, v interface{}) string {
	return key + "=" + anyToString(v)
}

func intToString(i int) string {
	if i == 0 {
		return "0"
	}

	neg := i < 0
	if neg {
		i = -i
	}

	buf := make([]byte, 0, 10)
	for i > 0 {
		buf = append(buf, byte('0'+i%10))
		i /= 10
	}

	if neg {
		buf = append(buf, '-')
	}

	// Reverse
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	return string(buf)
}

func anyToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return intToString(val)
	default:
		return "?"
	}
}
