package config

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var Settings InitConfig
var DefaultProfileManager *FileProfileManager

type InitConfig struct {
	Profile ProfileConfig `mapstructure:"profile"`
	Logger  LogConfig     `mapstructure:"log"`
}

type ProfileConfig struct {
	Default string `mapstructure:"default"`
}

type LogConfig struct {
	ZeroConfig `mapstructure:",squash"`
	File       LumberConfig `mapstructure:",squash"`
}

type ZeroConfig struct {
	LogLevel  string `mapstructure:"level"`
	LogWriter string `mapstructure:"writer"`
}
type LumberConfig struct {
	Path       string `mapstructure:"filepath"`
	MaxSize    int    `mapstructure:"maxsize"`
	MaxBackups int    `mapstructure:"maxbackups"`
	MaxAge     int    `mapstructure:"maxage"`
	Compress   bool   `mapstructure:"compress"`
}

// init
func Init() {
	execPath, err := os.Executable()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get executable path")
		log.Info().Msg("Using Default Config")
	}
	execDir := filepath.Dir(execPath)
	log.Info().Msgf("Executable directory: %s", execDir)

	log.Info().Msg(execPath)
	viper.AddConfigPath(filepath.Join(execDir, "../../data/var/run/data-manager/config"))
	viper.AddConfigPath(filepath.Join(execDir, "./data/var/run/data-manager/config"))
	viper.AddConfigPath(filepath.Join(execDir, "./"))
	viper.AddConfigPath(filepath.Join(execDir, "./config/"))
	viper.SetConfigName("config")

	err = viper.ReadInConfig()
	if err != nil {
		log.Error().Err(err).Msg("Failed to read config file")
		return
	}

	err = viper.Unmarshal(&Settings)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to decode into struct")
	}
	log.Debug().Msgf("config params : %+v", Settings)
	log.Info().Str("loglevel", Settings.Logger.LogLevel).Msg("Logger initialized with loglevel")
}

// ConfigManager structure definition
type ConfigManager struct {
	DefaultProfile string
	ProfileManager ProfileManager
	configFilePath string
	mu             sync.Mutex
}

// NewConfigManager loads the config from the specified path
func NewConfigManager(configPath string) (*ConfigManager, error) {
	configFilePath := filepath.Join(configPath, "config.json")
	defaultProfile, err := loadDefaultProfile(configFilePath)
	if err != nil {
		return nil, err
	}

	profilePath := filepath.Join(configPath, "profile", "profile.json")

	return &ConfigManager{
		DefaultProfile: defaultProfile,
		ProfileManager: NewProfileManager(profilePath),
		configFilePath: configFilePath,
	}, nil
}

// loadDefaultProfile loads the default profile from the config file
func loadDefaultProfile(configFilePath string) (string, error) {
	file, err := os.Open(configFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	var config struct {
		DefaultProfile string `json:"defaultProfile"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return "", err
	}

	if config.DefaultProfile == "" {
		return "", errors.New("defaultProfile not set in config.json")
	}

	return config.DefaultProfile, nil
}

// CreateConfig creates a new config.json file with the given data
func (cm *ConfigManager) CreateConfig(configData map[string]interface{}) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	file, err := os.Create(cm.configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.MarshalIndent(configData, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

// ReadConfig reads the config.json file and returns the data
func (cm *ConfigManager) ReadConfig() (map[string]interface{}, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	file, err := os.Open(cm.configFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var configData map[string]interface{}
	err = json.Unmarshal(data, &configData)
	if err != nil {
		return nil, err
	}

	return configData, nil
}

// UpdateConfig updates the config.json file with the given data
func (cm *ConfigManager) UpdateConfig(updatedData map[string]interface{}) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Read the current data
	currentData, err := cm.ReadConfig()
	if err != nil {
		return err
	}

	// Merge the updated data
	for key, value := range updatedData {
		currentData[key] = value
	}

	// Write the merged data back to the file
	return cm.CreateConfig(currentData)
}

// DeleteConfig deletes the config.json file
func (cm *ConfigManager) DeleteConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	err := os.Remove(cm.configFilePath)
	if err != nil {
		return err
	}

	return nil
}

// GetDefaultCredentials returns the default profile credentials
func (cm *ConfigManager) GetDefaultCredentials(provider string) (interface{}, error) {
	return cm.ProfileManager.LoadCredentialsByProfile(cm.DefaultProfile, provider)
}
