package models

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	ErrModelNotFound = errors.New("model not found")
	ErrModelExists   = errors.New("model already exists")
)

// ModelManager manages all AI models
type ModelManager struct {
	models       map[string]Model
	configs      map[string]*ModelConfig
	defaultModel string
	configFile   string
	mu           sync.RWMutex
}

// NewModelManager creates a new model manager
func NewModelManager() *ModelManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	configDir := filepath.Join(homeDir, ".ai")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config directory: %v\n", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

	return &ModelManager{
		models:     make(map[string]Model),
		configs:    make(map[string]*ModelConfig),
		configFile: configFile,
	}
}

// Init initializes the model manager
func (m *ModelManager) Init() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load configuration file
	if err := m.loadConfig(); err != nil {
		return err
	}

	// Initialize all models based on configuration
	for name, config := range m.configs {
		// This will create different model instances based on model type
		// Simplified handling, implementing factory methods for each model type
		model, err := CreateModel(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create model %s: %v\n", name, err)
			continue
		}

		m.models[name] = model

		// If this model is marked as default enabled, set it as default
		if config.DefaultEnabled {
			m.defaultModel = name
		}
	}

	// If no default model but models exist, set the first one as default
	if m.defaultModel == "" && len(m.models) > 0 {
		for name := range m.models {
			m.defaultModel = name
			break
		}
	}

	return nil
}

// loadConfig loads the configuration file
func (m *ModelManager) loadConfig() error {
	if _, err := os.Stat(m.configFile); os.IsNotExist(err) {
		// Configuration file doesn't exist, create default configuration
		return m.saveConfig()
	}

	data, err := os.ReadFile(m.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var configs map[string]*ModelConfig
	if err := yaml.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	m.configs = configs
	return nil
}

// saveConfig saves configuration to file
func (m *ModelManager) saveConfig() error {
	data, err := yaml.Marshal(m.configs)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultModel gets the default model
func (m *ModelManager) GetDefaultModel() (Model, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.defaultModel == "" {
		return nil, errors.New("no default model set")
	}

	model, exists := m.models[m.defaultModel]
	if !exists {
		return nil, ErrModelNotFound
	}

	return model, nil
}

// GetModel gets a model by name
func (m *ModelManager) GetModel(name string) (Model, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	model, exists := m.models[name]
	if !exists {
		return nil, ErrModelNotFound
	}

	return model, nil
}

// AddModel adds a new model
func (m *ModelManager) AddModel(name, url, apiKey string, defaultEnabled bool, chatOptions *ChatOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[name]; exists {
		return ErrModelExists
	}

	// Create new model configuration
	config := &ModelConfig{
		Name:               name,
		URL:                url,
		APIKey:             apiKey,
		DefaultEnabled:     defaultEnabled,
		DefaultChatOptions: chatOptions,
	}

	// Create model instance
	model, err := CreateModel(config)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Store model and configuration
	m.models[name] = model
	m.configs[name] = config

	// If this is the first model or defaultEnabled is true, set it as default
	if len(m.configs) == 1 || defaultEnabled {
		m.defaultModel = name
	}

	// Save configuration
	return m.saveConfig()
}

// RemoveModel removes a model
func (m *ModelManager) RemoveModel(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[name]; !exists {
		return ErrModelNotFound
	}

	// Delete model and configuration
	delete(m.models, name)
	delete(m.configs, name)

	// Save configuration
	return m.saveConfig()
}

// SetDefaultModel sets the default model
func (m *ModelManager) SetDefaultModel(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[name]; !exists {
		return ErrModelNotFound
	}

	// Set new default model
	m.defaultModel = name

	// Save configuration
	return m.saveConfig()
}

// ListModels lists all models
func (m *ModelManager) ListModels() map[string]*ModelConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid external modification
	result := make(map[string]*ModelConfig, len(m.configs))
	for k, v := range m.configs {
		configCopy := *v
		result[k] = &configCopy
	}

	return result
}

// GetDefaultModelName gets the name of the default model
func (m *ModelManager) GetDefaultModelName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.defaultModel
}

// UpdateModelConfig updates a model's configuration
func (m *ModelManager) UpdateModelConfig(name string, config *ModelConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.configs[name]; !exists {
		return ErrModelNotFound
	}

	// Update configuration
	m.configs[name] = config

	// If DefaultEnabled is true, set as default model
	if config.DefaultEnabled {
		m.defaultModel = name
	}

	// Recreate model instance with new configuration
	model, err := CreateModel(config)
	if err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}

	// Update model instance
	m.models[name] = model

	// Save configuration
	return m.saveConfig()
}

// GetModelConfig gets a model's configuration
func (m *ModelManager) GetModelConfig(name string) (*ModelConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.configs[name]
	if !exists {
		return nil, ErrModelNotFound
	}

	// Return a copy to avoid external modification
	configCopy := *config
	return &configCopy, nil
}
