package hardware

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// execCommand allows mocking exec.Command
var execCommand = exec.Command
var lookPath = exec.LookPath

// GPUType represents the type of GPU available
type GPUType string

const (
	GPUNone   GPUType = "none"
	GPUNvidia GPUType = "nvidia"
	GPUAMD    GPUType = "amd"
	GPUIntel  GPUType = "intel"
)

// Capabilities represents the hardware capabilities of a worker
type Capabilities struct {
	HasNvidia     bool    `json:"has_nvidia"`
	HasAMD        bool    `json:"has_amd"`
	HasIntelQSV   bool    `json:"has_intel_qsv"`
	HasVAAPI      bool    `json:"has_vaapi"`
	CPUCount      int     `json:"cpu_count"`
	GPUType       GPUType `json:"gpu_type"`
	GPUName       string  `json:"gpu_name,omitempty"`
	MemoryMB      int     `json:"memory_mb,omitempty"`
	FFmpegVersion string  `json:"ffmpeg_version,omitempty"`
}

// Detect detects the hardware capabilities of the current system
func Detect() *Capabilities {
	caps := &Capabilities{
		CPUCount: runtime.NumCPU(),
		GPUType:  GPUNone,
	}

	// Check NVIDIA
	if gpuName := checkNvidia(); gpuName != "" {
		caps.HasNvidia = true
		caps.GPUType = GPUNvidia
		caps.GPUName = gpuName
	}

	// Check AMD
	if checkAMD() {
		caps.HasAMD = true
		if caps.GPUType == GPUNone {
			caps.GPUType = GPUAMD
		}
	}

	// Check Intel QSV
	if checkIntelQSV() {
		caps.HasIntelQSV = true
		if caps.GPUType == GPUNone {
			caps.GPUType = GPUIntel
		}
	}

	// Check VAAPI (Linux)
	caps.HasVAAPI = checkVAAPI()

	// Get FFmpeg version
	caps.FFmpegVersion = getFFmpegVersion()

	return caps
}

// checkNvidia checks for NVIDIA GPU and returns the GPU name
func checkNvidia() string {
	// Check if nvidia-smi exists
	if _, err := lookPath("nvidia-smi"); err != nil {
		return ""
	}

	// Try to get GPU name
	cmd := execCommand("nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
	output, err := cmd.Output()
	if err != nil {
		return "nvidia (unknown model)"
	}

	name := strings.TrimSpace(string(output))
	if name == "" {
		return "nvidia (unknown model)"
	}

	// Return first GPU if multiple
	lines := strings.Split(name, "\n")
	return strings.TrimSpace(lines[0])
}

// checkAMD checks for AMD GPU via ROCm or VAAPI
func checkAMD() bool {
	// Check for AMD ROCm
	if _, err := os.Stat("/opt/rocm"); err == nil {
		return true
	}

	// Check for AMD via VAAPI on Linux
	if _, err := os.Stat("/dev/dri/renderD128"); err == nil {
		// Check if it's AMD by looking at driver
		output, err := execCommand("vainfo", "--display", "drm", "--device", "/dev/dri/renderD128").CombinedOutput()
		if err == nil && strings.Contains(string(output), "AMD") {
			return true
		}
	}

	return false
}

// checkIntelQSV checks for Intel Quick Sync Video
func checkIntelQSV() bool {
	// Check for Intel iHD driver via VAAPI
	output, err := execCommand("vainfo").CombinedOutput()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "iHD") || strings.Contains(string(output), "i965")
}

// checkVAAPI checks for VAAPI support (Linux)
func checkVAAPI() bool {
	_, err := os.Stat("/dev/dri/renderD128")
	return err == nil
}

// getFFmpegVersion returns the installed FFmpeg version
func getFFmpegVersion() string {
	cmd := execCommand("ffmpeg", "-version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		// Parse "ffmpeg version X.X.X ..."
		parts := strings.Fields(lines[0])
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	return ""
}

// GetEncoderForGPU returns the appropriate encoder based on GPU and codec
func GetEncoderForGPU(gpuType GPUType, codec string) string {
	switch codec {
	case "h264":
		switch gpuType {
		case GPUNvidia:
			return "h264_nvenc"
		case GPUAMD:
			return "h264_amf"
		case GPUIntel:
			return "h264_qsv"
		default:
			return "libx264"
		}
	case "h265", "hevc":
		switch gpuType {
		case GPUNvidia:
			return "hevc_nvenc"
		case GPUAMD:
			return "hevc_amf"
		case GPUIntel:
			return "hevc_qsv"
		default:
			return "libx265"
		}
	case "av1":
		switch gpuType {
		case GPUNvidia:
			return "av1_nvenc" // RTX 40+ only
		case GPUAMD:
			return "av1_amf" // RX 7000+ only
		case GPUIntel:
			return "av1_qsv" // Arc only
		default:
			return "libsvtav1"
		}
	default:
		return codec
	}
}

// SupportsHardwareEncoding returns true if any hardware encoding is available
func (c *Capabilities) SupportsHardwareEncoding() bool {
	return c.HasNvidia || c.HasAMD || c.HasIntelQSV
}

// GetBestEncoder returns the best encoder for the given codec
func (c *Capabilities) GetBestEncoder(codec string) string {
	return GetEncoderForGPU(c.GPUType, codec)
}
