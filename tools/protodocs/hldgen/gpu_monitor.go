package hldgen

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// GPUMonitor checks GPU availability and usage via nvidia-smi
type GPUMonitor struct {
	nvidiaSMIPath string
}

// NewGPUMonitor creates a new GPU monitor
func NewGPUMonitor() *GPUMonitor {
	return &GPUMonitor{
		nvidiaSMIPath: "nvidia-smi", // Assumes nvidia-smi is in PATH
	}
}

// GPUMetrics represents GPU utilization metrics
type GPUMetrics struct {
	UtilizationPercent int
	MemoryUsedMB       int
	MemoryTotalMB      int
	MemoryAvailableMB  int
	MemoryUsagePercent float64
}

// GetMetrics queries GPU status via nvidia-smi
func (m *GPUMonitor) GetMetrics() (*GPUMetrics, error) {
	cmd := exec.Command(m.nvidiaSMIPath,
		"--query-gpu=utilization.gpu,memory.used,memory.total",
		"--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("nvidia-smi command failed (GPU not available?): %w", err)
	}

	// Parse output: "85, 38912, 49152" (util%, used MB, total MB)
	parts := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(parts) != 3 {
		return nil, fmt.Errorf("unexpected nvidia-smi output format: %s", string(output))
	}

	util, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("failed to parse GPU utilization: %w", err)
	}

	used, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("failed to parse memory used: %w", err)
	}

	total, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil {
		return nil, fmt.Errorf("failed to parse memory total: %w", err)
	}

	available := total - used
	usagePercent := 0.0
	if total > 0 {
		usagePercent = float64(used) / float64(total) * 100.0
	}

	return &GPUMetrics{
		UtilizationPercent: util,
		MemoryUsedMB:       used,
		MemoryTotalMB:      total,
		MemoryAvailableMB:  available,
		MemoryUsagePercent: usagePercent,
	}, nil
}

// IsAvailable checks if GPU is ready for Ollama workloads
func (m *GPUMonitor) IsAvailable() bool {
	metrics, err := m.GetMetrics()
	if err != nil {
		// GPU monitoring failed - assume not available
		return false
	}

	// GPU available if:
	// 1. Utilization < 80% (not overloaded)
	// 2. At least 10GB free memory (enough for 70B model)
	return metrics.UtilizationPercent < 80 &&
		metrics.MemoryAvailableMB > 10000
}

// CanRunModel checks if GPU has enough memory for a specific model size
func (m *GPUMonitor) CanRunModel(modelSizeGB int) (bool, error) {
	metrics, err := m.GetMetrics()
	if err != nil {
		return false, err
	}

	// Convert GB to MB
	requiredMB := modelSizeGB * 1024

	// Need some overhead (~20%) for VRAM allocation
	requiredMB = int(float64(requiredMB) * 1.2)

	return metrics.MemoryAvailableMB >= requiredMB, nil
}

// String returns a human-readable representation of GPU status
func (m *GPUMetrics) String() string {
	return fmt.Sprintf(
		"GPU: %d%% util, Memory: %dMB / %dMB (%.1f%% used, %dMB available)",
		m.UtilizationPercent,
		m.MemoryUsedMB,
		m.MemoryTotalMB,
		m.MemoryUsagePercent,
		m.MemoryAvailableMB,
	)
}

// ShouldUseOllama determines if Ollama should be used based on GPU availability
// This can be used in LLMRouter for cost-aware provider selection
func (m *GPUMonitor) ShouldUseOllama() bool {
	if !m.IsAvailable() {
		return false
	}

	metrics, err := m.GetMetrics()
	if err != nil {
		return false
	}

	// Use Ollama if GPU is available and not heavily loaded
	// This saves cloud API costs while maintaining good performance
	return metrics.UtilizationPercent < 70 &&
		metrics.MemoryAvailableMB > 15000 // Conservative threshold
}
