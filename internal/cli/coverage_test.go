package cli

import (
	"errors"
	"testing"

	"go.uber.org/zap"
)

func TestErrors(t *testing.T) {
	if ErrConfigNotLoaded == nil {
		t.Error("ErrConfigNotLoaded should not be nil")
	}
	if ErrConfigNotLoaded.Error() != "配置未加载" {
		t.Errorf("ErrConfigNotLoaded.Error() = %q, want %q", ErrConfigNotLoaded.Error(), "配置未加载")
	}

	if ErrRegistryNotInitialized == nil {
		t.Error("ErrRegistryNotInitialized should not be nil")
	}
	if ErrRegistryNotInitialized.Error() != "注册中心未初始化" {
		t.Errorf("ErrRegistryNotInitialized.Error() = %q, want %q", ErrRegistryNotInitialized.Error(), "注册中心未初始化")
	}

	if !errors.Is(ErrConfigNotLoaded, ErrConfigNotLoaded) {
		t.Error("errors.Is should return true for same error")
	}
}

func TestGetConfigFile(t *testing.T) {
	original := cfgFile
	defer func() { cfgFile = original }()

	cfgFile = ""
	if got := GetConfigFile(); got != "" {
		t.Errorf("GetConfigFile() = %q, want empty string", got)
	}

	cfgFile = "/path/to/config.yaml"
	if got := GetConfigFile(); got != "/path/to/config.yaml" {
		t.Errorf("GetConfigFile() = %q, want %q", got, "/path/to/config.yaml")
	}
}

func TestInitLogger(t *testing.T) {
	originalLogger := logger
	originalSugar := sugar
	defer func() {
		logger = originalLogger
		sugar = originalSugar
	}()

	err := InitLogger(false)
	if err != nil {
		t.Errorf("InitLogger(false) error = %v", err)
	}

	if logger == nil {
		t.Error("InitLogger should set logger")
	}

	if sugar == nil {
		t.Error("InitLogger should set sugar")
	}

	_ = Sync()
}

func TestInitLoggerVerbose(t *testing.T) {
	originalLogger := logger
	originalSugar := sugar
	defer func() {
		logger = originalLogger
		sugar = originalSugar
	}()

	err := InitLogger(true)
	if err != nil {
		t.Errorf("InitLogger(true) error = %v", err)
	}

	if logger == nil {
		t.Error("InitLogger(true) should set logger")
	}

	_ = Sync()
}

func TestGetLogger(t *testing.T) {
	originalLogger := logger
	defer func() { logger = originalLogger }()

	logger = nil
	if got := GetLogger(); got != nil {
		t.Errorf("GetLogger() with nil logger = %v, want nil", got)
	}

	l := zap.NewNop()
	logger = l
	if got := GetLogger(); got == nil {
		t.Error("GetLogger() should return non-nil when logger is set")
	}
}

func TestGetSugar(t *testing.T) {
	originalSugar := sugar
	defer func() { sugar = originalSugar }()

	sugar = nil
	if got := GetSugar(); got != nil {
		t.Errorf("GetSugar() with nil sugar = %v, want nil", got)
	}

	s := zap.NewNop().Sugar()
	sugar = s
	if got := GetSugar(); got == nil {
		t.Error("GetSugar() should return non-nil when sugar is set")
	}
}

func TestSync(t *testing.T) {
	originalLogger := logger
	defer func() { logger = originalLogger }()

	logger = nil
	if err := Sync(); err != nil {
		t.Errorf("Sync() with nil logger error = %v", err)
	}

	l := zap.NewNop()
	logger = l
	if err := Sync(); err != nil {
		t.Errorf("Sync() with valid logger error = %v", err)
	}
}

func TestSetVersion(t *testing.T) {
	originalVersion := Version
	originalGitCommit := GitCommit
	originalBuildDate := BuildDate
	defer func() {
		Version = originalVersion
		GitCommit = originalGitCommit
		BuildDate = originalBuildDate
	}()

	SetVersion("1.0.0", "abc123", "2024-01-01")
	if Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", Version, "1.0.0")
	}
	if GitCommit != "abc123" {
		t.Errorf("GitCommit = %q, want %q", GitCommit, "abc123")
	}
	if BuildDate != "2024-01-01" {
		t.Errorf("BuildDate = %q, want %q", BuildDate, "2024-01-01")
	}

	SetVersion("", "", "")
	if Version != "1.0.0" {
		t.Errorf("SetVersion with empty strings should not change values, got %q", Version)
	}
}

func TestRootCmdPersistentFlags(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	cfgFlag := rootCmd.Flag("config")
	if cfgFlag == nil {
		t.Error("rootCmd should have 'config' flag")
	}

	verboseFlag := rootCmd.Flag("verbose")
	if verboseFlag == nil {
		t.Error("rootCmd should have 'verbose' flag")
	}

	tokenFlag := rootCmd.Flag("token")
	if tokenFlag == nil {
		t.Error("rootCmd should have 'token' flag")
	}
}
