package config

// Config represents the main configuration structure for the application
type Config struct {
	Application *Application `mapstructure:"application"`
	Services    *Services    `mapstructure:"services"`
}

// Interface implementation for Config
func (c *Config) GetApplication() ApplicationProvider {
	return c.Application
}

func (c *Config) GetServices() ServicesProvider {
	return c.Services
}
