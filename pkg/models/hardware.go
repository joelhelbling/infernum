package models

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
)

type HardwareInfo struct {
	OSName       string  `json:"os_name"`
	OSVersion    string  `json:"os_version"`
	Architecture string  `json:"architecture"`
	CPUModel     string  `json:"cpu_model"`
	CPUCores     int     `json:"cpu_cores"`
	RAMGB        float64 `json:"ram_gb"`
	GPUName      string  `json:"gpu_name,omitempty"`
	VRAMGB       float64 `json:"vram_gb,omitempty"`
}

// Fingerprint computes a SHA-256 hash of normalized hardware fields.
// RAM and VRAM are rounded to the nearest GB to avoid OS reporting variance.
func (h HardwareInfo) Fingerprint() string {
	ramRounded := int(math.Round(h.RAMGB))
	vramRounded := int(math.Round(h.VRAMGB))

	parts := []string{
		strings.ToLower(h.OSName),
		strings.ToLower(h.OSVersion),
		strings.ToLower(h.Architecture),
		strings.ToLower(h.CPUModel),
		fmt.Sprintf("%d", h.CPUCores),
		fmt.Sprintf("%d", ramRounded),
		strings.ToLower(h.GPUName),
		fmt.Sprintf("%d", vramRounded),
	}

	input := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}
