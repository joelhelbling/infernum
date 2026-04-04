package models_test

import (
	"testing"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestHardwareFingerprint(t *testing.T) {
	hw := models.HardwareInfo{
		OSName:       "linux",
		OSVersion:    "6.17.0",
		Architecture: "x86_64",
		CPUModel:     "AMD Ryzen 9 7950X",
		CPUCores:     32,
		RAMGB:        63.9, // should round to 64
		GPUName:      "NVIDIA RTX 4090",
		VRAMGB:       24.0,
	}

	fp1 := hw.Fingerprint()
	if fp1 == "" {
		t.Fatal("fingerprint should not be empty")
	}

	// Same specs, slightly different RAM reporting — should produce same fingerprint
	hw2 := hw
	hw2.RAMGB = 64.1
	fp2 := hw2.Fingerprint()
	if fp1 != fp2 {
		t.Errorf("fingerprints should match after rounding: %q != %q", fp1, fp2)
	}

	// Different GPU — should produce different fingerprint
	hw3 := hw
	hw3.GPUName = "NVIDIA RTX 3090"
	fp3 := hw3.Fingerprint()
	if fp1 == fp3 {
		t.Error("fingerprints should differ for different GPUs")
	}
}

func TestHardwareFingerprintCaseInsensitive(t *testing.T) {
	hw1 := models.HardwareInfo{
		OSName:       "Linux",
		Architecture: "x86_64",
		CPUModel:     "AMD Ryzen 9 7950X",
		CPUCores:     32,
		RAMGB:        64,
	}
	hw2 := hw1
	hw2.OSName = "linux"

	if hw1.Fingerprint() != hw2.Fingerprint() {
		t.Error("fingerprint should be case-insensitive")
	}
}
