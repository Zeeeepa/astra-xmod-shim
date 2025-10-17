package config

import (
	"fmt"
	"os"
	"sync"

	confSpec "astron-xmod-shim/internal/dto/config"

	"github.com/spf13/viper"
)

var (
	globalConfig *confSpec.GlobalConfig
	once         sync.Once
	initErr      error
	configPath   string
)

// SetConfigPath Set config file path in advance
func SetConfigPath(path string) {
	configPath = path
}

// Get Lazily load config instance (thread-safe)
func Get() *confSpec.GlobalConfig {
	once.Do(func() {
		if configPath == "" {
			initErr = fmt.Errorf("config path not set")
			return
		}

		v := viper.New()
		v.SetConfigFile(configPath)
		v.SetConfigType("yaml")

		if err := v.ReadInConfig(); err != nil {
			initErr = fmt.Errorf("failed to read config file: %w", err)
			return
		}

		globalConfig = &confSpec.GlobalConfig{}
		if err := v.Unmarshal(globalConfig); err != nil {
			initErr = fmt.Errorf("failed to parse config: %w", err)
			globalConfig = nil
			return
		}
	})

	// Return nil if loading failed, as expected
	if initErr != nil {
		// Optional: Print error log or expose initErr through other means
		// log.Printf("config loading failed: %v", initErr)
		return nil
	}

	return globalConfig
}

// GetConfFromFileDir Load specific type of config from specified file path
// Supports loading YAML format config files
// configPath: Full path to config file
// Returns loaded config instance
func GetConfFromFileDir[T any](configPath string) (*T, error) {
	// Create new instance of T and get its pointer
	conf := new(T)
	// Check if file exists
	stat, err := os.Stat(configPath)
	if err != nil {
		return nil, fmt.Errorf("config file does not exist: %w", err)
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("path is not a file: %s", configPath)
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config into newly created struct pointer
	if err := v.Unmarshal(conf); err != nil {
		return nil, fmt.Errorf("failed to parse config into struct: %w", err)
	}

	return conf, nil
}