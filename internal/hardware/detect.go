package hardware

import (
	"math"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func Detect() (models.HardwareInfo, error) {
	info := models.HardwareInfo{
		Architecture: runtime.GOARCH,
		CPUCores:     runtime.NumCPU(),
	}

	detectOS(&info)
	detectCPU(&info)
	detectRAM(&info)
	detectGPU(&info)

	return info, nil
}

func detectOS(info *models.HardwareInfo) {
	switch runtime.GOOS {
	case "linux":
		info.OSName = "linux"
		if out, err := exec.Command("uname", "-r").Output(); err == nil {
			info.OSVersion = strings.TrimSpace(string(out))
		}
	case "darwin":
		info.OSName = "macos"
		if out, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
			info.OSVersion = strings.TrimSpace(string(out))
		}
	case "windows":
		info.OSName = "windows"
		if out, err := exec.Command("cmd", "/c", "ver").Output(); err == nil {
			info.OSVersion = strings.TrimSpace(string(out))
		}
	default:
		info.OSName = runtime.GOOS
	}
}

func detectCPU(info *models.HardwareInfo) {
	switch runtime.GOOS {
	case "linux":
		if out, err := exec.Command("sh", "-c", `grep -m1 'model name' /proc/cpuinfo | cut -d: -f2`).Output(); err == nil {
			info.CPUModel = strings.TrimSpace(string(out))
		}
	case "darwin":
		if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
			info.CPUModel = strings.TrimSpace(string(out))
		}
	}
	if info.CPUModel == "" {
		info.CPUModel = "unknown"
	}
}

func detectRAM(info *models.HardwareInfo) {
	switch runtime.GOOS {
	case "linux":
		if out, err := exec.Command("sh", "-c", `grep MemTotal /proc/meminfo | awk '{print $2}'`).Output(); err == nil {
			if kb, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64); err == nil {
				info.RAMGB = math.Round(kb/1024/1024*10) / 10
			}
		}
	case "darwin":
		if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
			if bytes, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64); err == nil {
				info.RAMGB = math.Round(bytes/1024/1024/1024*10) / 10
			}
		}
	}
}

func detectGPU(info *models.HardwareInfo) {
	// Try NVIDIA first
	if out, err := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader,nounits").Output(); err == nil {
		name := strings.TrimSpace(strings.Split(string(out), "\n")[0])
		if name != "" {
			info.GPUName = name
			detectNvidiaVRAM(info)
			return
		}
	}

	// Try Apple Silicon
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		info.GPUName = "Apple Silicon (integrated)"
		// On Apple Silicon, GPU shares system memory — VRAM is not separately reportable
		return
	}

	// Try AMD via rocm-smi
	if out, err := exec.Command("rocm-smi", "--showproductname").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Card") {
				fields := strings.Fields(line)
				if len(fields) > 2 {
					info.GPUName = strings.Join(fields[2:], " ")
					break
				}
			}
		}
	}
}

func detectNvidiaVRAM(info *models.HardwareInfo) {
	if out, err := exec.Command("nvidia-smi", "--query-gpu=memory.total", "--format=csv,noheader,nounits").Output(); err == nil {
		if mb, err := strconv.ParseFloat(strings.TrimSpace(strings.Split(string(out), "\n")[0]), 64); err == nil {
			info.VRAMGB = math.Round(mb/1024*10) / 10
		}
	}
}
