// log_test.go
package log

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	config "astron-xmod-shim/internal/dto/config"

	"github.com/stretchr/testify/assert"
)

// resetGlobalLogger 重置全局日志器，防止测试间污染
func resetGlobalLogger() {
	globalLogger = nil
	sugarLogger = nil
}

func TestInit(t *testing.T) {
	resetGlobalLogger()
	defer resetGlobalLogger()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdout = w

	cfg := &config.LogConfig{
		Level:    "info",
		ShowLine: true,
	}

	err = Init(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, globalLogger)
	assert.NotNil(t, sugarLogger)

	Info("test init message")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if output != "" {
		var logEntry map[string]interface{}
		err := json.Unmarshal([]byte(output), &logEntry)
		if assert.NoError(t, err, "Failed to parse JSON: %s", output) {
			assert.Contains(t, logEntry, "level")
			assert.Contains(t, logEntry, "time")
			assert.Contains(t, logEntry, "msg")
			assert.Contains(t, logEntry, "caller")
		}
	}
}

func TestInitWithInvalidLevel(t *testing.T) {
	resetGlobalLogger()
	defer resetGlobalLogger()

	cfg := &config.LogConfig{
		Level:    "invalid-level",
		ShowLine: false,
	}

	err := Init(cfg)
	assert.NoError(t, err) // 应默认降级为 info
}

func TestSetDefaultConfig(t *testing.T) {
	cfg := &config.LogConfig{}
	err := setDefaultConfig(cfg)
	assert.NoError(t, err)
	assert.Equal(t, "info", cfg.Level)
}

func TestBuildLogCore(t *testing.T) {
	cfg := &config.LogConfig{Level: "debug"}
	core := buildLogCore(cfg)
	assert.NotNil(t, core)
}

func TestBuildZapOptions(t *testing.T) {
	cfg := &config.LogConfig{ShowLine: true}
	options := buildZapOptions(cfg)
	assert.Len(t, options, 2)

	cfg2 := &config.LogConfig{ShowLine: false}
	options2 := buildZapOptions(cfg2)
	assert.Empty(t, options2)
}

func TestLogFunctions(t *testing.T) {
	resetGlobalLogger()
	defer resetGlobalLogger()

	cfg := &config.LogConfig{Level: "debug", ShowLine: false}
	err := Init(cfg)
	assert.NoError(t, err)

	assert.NotPanics(t, func() {
		Debug("debug message %s", "test")
		Info("info message %s", "test")
		Warn("warn message %s", "test")
		Error("error message %s", "test")
	})
}

func TestFatal(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		resetGlobalLogger()
		cfg := &config.LogConfig{Level: "debug"}
		_ = Init(cfg)
		Fatal("fatal test message")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatal")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	exitErr, ok := err.(*exec.ExitError)
	if assert.True(t, ok, "expected process to exit") {
		assert.False(t, exitErr.Success(), "expected non-zero exit code")
	}
}

func TestSync(t *testing.T) {
	resetGlobalLogger()
	defer resetGlobalLogger()

	cfg := &config.LogConfig{Level: "info"}
	err := Init(cfg)
	assert.NoError(t, err)

	// Sync 可能因 stdout 是 pipe 而报错（如 "bad file descriptor"）
	// 在云原生环境中，stdout 不需要 sync，因此我们只验证不 panic
	assert.NotPanics(t, func() {
		_ = Sync() // 忽略返回值，重点是不 panic
	})

	// 测试未初始化时调用 Sync
	globalLogger = nil
	assert.NotPanics(t, func() {
		_ = Sync()
	})
}

func TestLogOutputFormat(t *testing.T) {
	resetGlobalLogger()
	defer resetGlobalLogger()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdout = w

	cfg := &config.LogConfig{Level: "info", ShowLine: true}
	err = Init(cfg)
	assert.NoError(t, err)

	Info("test message %s", "value")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "test message value")
	assert.Contains(t, output, `"level":"info"`)
	assert.Contains(t, output, `"time":"`)
	assert.Contains(t, output, `"caller":"`)

	var logEntry map[string]interface{}
	err = json.Unmarshal([]byte(output), &logEntry)
	assert.NoError(t, err, "Invalid JSON: %s", output)
}

func TestMultipleInitCalls(t *testing.T) {
	resetGlobalLogger()
	defer resetGlobalLogger()

	cfg1 := &config.LogConfig{Level: "debug"}
	err := Init(cfg1)
	assert.NoError(t, err)

	cfg2 := &config.LogConfig{Level: "error"}
	err = Init(cfg2)
	assert.NoError(t, err)
	assert.NotNil(t, globalLogger)
}
