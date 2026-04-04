package hardware_test

import (
	"testing"

	"github.com/joelhelbling/ollama-bench/internal/hardware"
)

func TestDetectReturnsPopulatedFields(t *testing.T) {
	info, err := hardware.Detect()
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	if info.OSName == "" {
		t.Error("OSName should not be empty")
	}
	if info.Architecture == "" {
		t.Error("Architecture should not be empty")
	}
	if info.CPUModel == "" {
		t.Error("CPUModel should not be empty")
	}
	if info.CPUCores <= 0 {
		t.Errorf("CPUCores should be positive, got %d", info.CPUCores)
	}
	if info.RAMGB <= 0 {
		t.Errorf("RAMGB should be positive, got %f", info.RAMGB)
	}
}

func TestDetectFingerprint(t *testing.T) {
	info, err := hardware.Detect()
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	fp := info.Fingerprint()
	if fp == "" {
		t.Error("Fingerprint should not be empty")
	}

	// Running twice should produce the same fingerprint
	info2, _ := hardware.Detect()
	if info.Fingerprint() != info2.Fingerprint() {
		t.Error("Fingerprint should be stable across calls")
	}
}
