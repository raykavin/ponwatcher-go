package viper

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// LoaderOptions contains configuration options for the loader
type LoaderOptions[T any] struct {
	ConfigName          string            // Config file name (without extension)
	ConfigType          string            // Config file type (yaml, json, toml, etc.)
	ConfigPaths         []string          // Paths to search for config file
	EnvPrefix           string            // Environment variable prefix
	EnvKeyReplacer      *strings.Replacer // Custom replacer for env keys
	AutomaticEnv        bool              // Enable automatic environment variable binding
	WatchConfig         bool              // Enable configuration file watching
	ReloadDebounce      time.Duration     // Debounce duration for config reloads
	OnConfigChange      func(*T)          // Callback when config changes (optional)
	OnConfigChangeError func(error)       // Callback when config reload fails (optional)
}

// ConfigChangeEvent represents a configuration change event
type ConfigChangeEvent[T any] struct {
	OldConfig *T
	NewConfig *T
	Error     error
	Timestamp time.Time
}

// ConfigWatcher defines the interface for configuration change notifications
type ConfigWatcher[T any] interface {
	Subscribe() <-chan ConfigChangeEvent[T]
	Unsubscribe(<-chan ConfigChangeEvent[T])
}

// DefaultLoaderOptions returns sensible defaults for configuration loading
func DefaultLoaderOptions[T any]() *LoaderOptions[T] {
	return &LoaderOptions[T]{
		ConfigName:          "config",
		ConfigType:          "yaml",
		ConfigPaths:         []string{".", "./config", "/etc/app", "$HOME/.app"},
		EnvPrefix:           "APP",
		EnvKeyReplacer:      strings.NewReplacer(".", "_", "-", "_"),
		AutomaticEnv:        true,
		WatchConfig:         false,
		ReloadDebounce:      1 * time.Second,
		OnConfigChange:      nil,
		OnConfigChangeError: nil,
	}
}

// ConfigLoader provides methods for loading configuration
type ConfigLoader[T any] struct {
	options     *LoaderOptions[T]
	viper       *viper.Viper
	current     *T
	mutex       sync.RWMutex
	subscribers []chan ConfigChangeEvent[T]
	subMutex    sync.RWMutex
	validator   func(*T) error
	ctx         context.Context
	cancel      context.CancelFunc
	debouncer   *time.Timer
}

// New creates a new configuration loader with the given options
func New[T any](options *LoaderOptions[T]) *ConfigLoader[T] {
	if options == nil {
		options = DefaultLoaderOptions[T]()
	}

	v := viper.New()

	// Set config file properties
	v.SetConfigName(options.ConfigName)
	v.SetConfigType(options.ConfigType)

	// Add config paths
	for _, path := range options.ConfigPaths {
		v.AddConfigPath(path)
	}

	// Set environment variable properties
	if options.AutomaticEnv {
		v.AutomaticEnv()
	}

	if options.EnvPrefix != "" {
		v.SetEnvPrefix(options.EnvPrefix)
	}

	if options.EnvKeyReplacer != nil {
		v.SetEnvKeyReplacer(options.EnvKeyReplacer)
	}

	ctx, cancel := context.WithCancel(context.Background())

	cl := &ConfigLoader[T]{
		options:     options,
		viper:       v,
		ctx:         ctx,
		cancel:      cancel,
		subscribers: make([]chan ConfigChangeEvent[T], 0),
	}

	return cl
}

// Load reads configuration from file and environment variables
func (cl *ConfigLoader[T]) Load() (*T, error) {
	config, err := cl.loadConfig()
	if err != nil {
		return nil, err
	}

	cl.mutex.Lock()
	cl.current = config
	cl.mutex.Unlock()

	// Start watching for changes if enabled
	if cl.options.WatchConfig {
		cl.startWatching()
	}

	return config, nil
}

// loadConfig performs the actual configuration loading
func (cl *ConfigLoader[T]) loadConfig() (*T, error) {
	var config T

	// Try to read config file
	if err := cl.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, continue with env vars only
			fmt.Printf("Config file not found, using environment variables only\n")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		fmt.Printf("Using config file: %s\n", cl.viper.ConfigFileUsed())
	}

	// Bind environment variables based on struct tags
	if err := cl.bindEnvVars(reflect.TypeOf(config), ""); err != nil {
		return nil, fmt.Errorf("error binding environment variables: %w", err)
	}

	// Unmarshal configuration
	if err := cl.viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling configuration: %w", err)
	}

	return &config, nil
}

// LoadWithValidation loads configuration and runs validation
func (cl *ConfigLoader[T]) LoadWithValidation(validator func(*T) error) (*T, error) {
	cl.validator = validator

	config, err := cl.Load()
	if err != nil {
		return nil, err
	}

	if validator != nil {
		if err := validator(config); err != nil {
			return nil, fmt.Errorf("configuration validation failed: %w", err)
		}
	}

	return config, nil
}

// GetCurrent returns the current configuration (thread-safe)
func (cl *ConfigLoader[T]) GetCurrent() *T {
	cl.mutex.RLock()
	defer cl.mutex.RUnlock()
	return cl.current
}

// Reload manually reloads the configuration
func (cl *ConfigLoader[T]) Reload() error {
	oldConfig := cl.GetCurrent()

	newConfig, err := cl.loadConfig()
	if err != nil {
		cl.notifyError(err)
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	// Validate new configuration
	if cl.validator != nil {
		if err := cl.validator(newConfig); err != nil {
			validationErr := fmt.Errorf("configuration validation failed during reload: %w", err)
			cl.notifyError(validationErr)
			return validationErr
		}
	}

	cl.mutex.Lock()
	cl.current = newConfig
	cl.mutex.Unlock()

	// Notify subscribers
	cl.notifyChange(oldConfig, newConfig)

	return nil
}

// startWatching starts watching for configuration file changes
func (cl *ConfigLoader[T]) startWatching() {
	cl.viper.WatchConfig()
	cl.viper.OnConfigChange(func(e fsnotify.Event) {
		cl.debouncedReload()
	})
}

// debouncedReload implements debounced configuration reloading
func (cl *ConfigLoader[T]) debouncedReload() {
	if cl.debouncer != nil {
		cl.debouncer.Stop()
	}

	cl.debouncer = time.AfterFunc(cl.options.ReloadDebounce, func() {
		if err := cl.Reload(); err != nil {
			fmt.Printf("Error reloading configuration: %v\n", err)
		}
	})
}

// Subscribe returns a channel that receives configuration change events
func (cl *ConfigLoader[T]) Subscribe() <-chan ConfigChangeEvent[T] {
	cl.subMutex.Lock()
	defer cl.subMutex.Unlock()

	ch := make(chan ConfigChangeEvent[T], 10) // Buffered channel
	cl.subscribers = append(cl.subscribers, ch)
	return ch
}

// Unsubscribe removes a subscription channel
func (cl *ConfigLoader[T]) Unsubscribe(ch <-chan ConfigChangeEvent[T]) {
	cl.subMutex.Lock()
	defer cl.subMutex.Unlock()

	for i, subscriber := range cl.subscribers {
		if subscriber == ch {
			close(subscriber)
			cl.subscribers = append(cl.subscribers[:i], cl.subscribers[i+1:]...)
			break
		}
	}
}

// notifyChange notifies all subscribers about configuration changes
func (cl *ConfigLoader[T]) notifyChange(oldConfig, newConfig *T) {
	event := ConfigChangeEvent[T]{
		OldConfig: oldConfig,
		NewConfig: newConfig,
		Timestamp: time.Now(),
	}

	// Call registered callback if provided
	if cl.options.OnConfigChange != nil {
		go cl.options.OnConfigChange(newConfig)
	}

	// Notify subscribers
	cl.subMutex.RLock()
	defer cl.subMutex.RUnlock()

	for _, subscriber := range cl.subscribers {
		select {
		case subscriber <- event:
		default:
			// Channel is full, skip this subscriber
			fmt.Printf("Warning: Configuration change subscriber channel is full\n")
		}
	}
}

// notifyError notifies about configuration reload errors
func (cl *ConfigLoader[T]) notifyError(err error) {
	event := ConfigChangeEvent[T]{
		Error:     err,
		Timestamp: time.Now(),
	}

	// Call registered error callback if provided
	if cl.options.OnConfigChangeError != nil {
		go cl.options.OnConfigChangeError(err)
	}

	// Notify subscribers about error
	cl.subMutex.RLock()
	defer cl.subMutex.RUnlock()

	for _, subscriber := range cl.subscribers {
		select {
		case subscriber <- event:
		default:
			fmt.Printf("Warning: Configuration error subscriber channel is full\n")
		}
	}
}

// Stop stops the configuration loader and cleans up resources
func (cl *ConfigLoader[T]) Stop() {
	cl.cancel()

	if cl.debouncer != nil {
		cl.debouncer.Stop()
	}

	// Close all subscriber channels
	cl.subMutex.Lock()
	defer cl.subMutex.Unlock()

	for _, subscriber := range cl.subscribers {
		close(subscriber)
	}
	cl.subscribers = nil
}

// bindEnvVars recursively binds environment variables based on mapstructure tags
func (cl *ConfigLoader[T]) bindEnvVars(t reflect.Type, prefix string) error {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue
		}

		// Handle squash tag for embedded structs
		if strings.Contains(tag, "squash") {
			if err := cl.bindEnvVars(field.Type, prefix); err != nil {
				return err
			}
			continue
		}

		// Extract the actual field name from tag (handle omitempty, etc.)
		fieldName := strings.Split(tag, ",")[0]
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}

		var envKey string
		if prefix != "" {
			envKey = prefix + "." + fieldName
		} else {
			envKey = fieldName
		}

		// Bind the environment variable
		if err := cl.viper.BindEnv(envKey); err != nil {
			return fmt.Errorf("failed to bind env var for %s: %w", envKey, err)
		}

		// Recursively handle nested structs
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct {
			if err := cl.bindEnvVars(fieldType, envKey); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetViper returns the underlying viper instance for advanced usage
func (cl *ConfigLoader[T]) GetViper() *viper.Viper {
	return cl.viper
}

// SetConfigValue sets a configuration value programmatically
func (cl *ConfigLoader[T]) SetConfigValue(key string, value interface{}) {
	cl.viper.Set(key, value)
}

// GetConfigValue gets a configuration value
func (cl *ConfigLoader[T]) GetConfigValue(key string) interface{} {
	return cl.viper.Get(key)
}

// PrintEnvHelp prints all the environment variables that can be used
func (cl *ConfigLoader[T]) PrintEnvHelp() {
	var config T
	fmt.Println("Environment Variables:")
	fmt.Println("=====================")
	cl.printEnvHelp(reflect.TypeOf(config), "", 0)
}

// printEnvHelp recursively prints environment variable help
func (cl *ConfigLoader[T]) printEnvHelp(t reflect.Type, prefix string, depth int) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	indent := strings.Repeat("  ", depth)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue
		}

		if strings.Contains(tag, "squash") {
			cl.printEnvHelp(field.Type, prefix, depth)
			continue
		}

		fieldName := strings.Split(tag, ",")[0]
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}

		var envKey string
		if prefix != "" {
			envKey = prefix + "." + fieldName
		} else {
			envKey = fieldName
		}

		// Convert to environment variable format
		envVar := cl.options.EnvPrefix + "_" + strings.ToUpper(cl.options.EnvKeyReplacer.Replace(envKey))

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct {
			fmt.Printf("%s%s (%s):\n", indent, envVar, fieldType.Name())
			cl.printEnvHelp(fieldType, envKey, depth+1)
		} else {
			fmt.Printf("%s%s (%s)\n", indent, envVar, fieldType.String())
		}
	}
}
