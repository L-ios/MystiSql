package validator

import (
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type ValidatorConfig struct {
	Whitelist []string `yaml:"whitelist"`
	Blacklist []string `yaml:"blacklist"`
}

type ConfigPersistence struct {
	configPath string
	logger     *zap.Logger
	mu         sync.RWMutex
}

func NewConfigPersistence(configPath string, logger *zap.Logger) *ConfigPersistence {
	return &ConfigPersistence{
		configPath: configPath,
		logger:     logger,
	}
}

func (cp *ConfigPersistence) Save(config *ValidatorConfig) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	dir := filepath.Dir(cp.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		cp.logger.Error("Failed to marshal config", zap.Error(err))
		return err
	}

	if err := os.WriteFile(cp.configPath, data, 0644); err != nil {
		cp.logger.Error("Failed to write config file", zap.Error(err))
		return err
	}

	cp.logger.Info("Validator config saved", zap.String("path", cp.configPath))
	return nil
}

func (cp *ConfigPersistence) Load() (*ValidatorConfig, error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	data, err := os.ReadFile(cp.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ValidatorConfig{
				Whitelist: []string{},
				Blacklist: []string{},
			}, nil
		}
		cp.logger.Error("Failed to read config file", zap.Error(err))
		return nil, err
	}

	var config ValidatorConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		cp.logger.Error("Failed to unmarshal config", zap.Error(err))
		return nil, err
	}

	if config.Whitelist == nil {
		config.Whitelist = []string{}
	}
	if config.Blacklist == nil {
		config.Blacklist = []string{}
	}

	return &config, nil
}
