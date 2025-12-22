package hardware

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test logic for execution mocking
var mockHelperProcessResponse = ""
var mockHelperProcessCmd = ""

func fakeExecCommand(name string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", name}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "MOCK_RESPONSE=" + mockHelperProcessResponse, "MOCK_CMD=" + mockHelperProcessCmd}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		os.Exit(2)
	}

	_ = args[0]
	// Check if this is the expected command
	// Simplified: just print response
	fmt.Print(os.Getenv("MOCK_RESPONSE"))
}

func TestIdentifyNvidia(t *testing.T) {
	// Mock lookPath
	oldLookPath := lookPath
	oldExec := execCommand
	defer func() {
		lookPath = oldLookPath
		execCommand = oldExec
	}()

	lookPath = func(file string) (string, error) {
		if file == "nvidia-smi" {
			return "/usr/bin/nvidia-smi", nil
		}
		return "", fmt.Errorf("not found")
	}

	execCommand = fakeExecCommand
	mockHelperProcessResponse = "Tesla T4\n"

	caps := Detect()
	assert.True(t, caps.HasNvidia)
	assert.Equal(t, "Tesla T4", caps.GPUName)
	assert.Equal(t, GPUNvidia, caps.GPUType)
}

func TestGetEncoderForGPU(t *testing.T) {
	tests := []struct {
		gpu    GPUType
		codec  string
		expect string
	}{
		{GPUNvidia, "h264", "h264_nvenc"},
		{GPUNvidia, "hevc", "hevc_nvenc"},
		{GPUAMD, "h264", "h264_amf"},
		{GPUIntel, "av1", "av1_qsv"},
		{GPUNone, "h264", "libx264"},
	}

	for _, tt := range tests {
		res := GetEncoderForGPU(tt.gpu, tt.codec)
		assert.Equal(t, tt.expect, res)
	}
}

func TestGetFFmpegVersion(t *testing.T) {
	oldExec := execCommand
	defer func() { execCommand = oldExec }()
	execCommand = fakeExecCommand
	mockHelperProcessResponse = "ffmpeg version 6.0-static ...\n"

	ver := getFFmpegVersion()
	assert.Equal(t, "6.0-static", ver)
}
