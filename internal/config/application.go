package config

// Application contains application-level configuration settings
type Application struct {
	Name        string     `mapstructure:"name"`
	Description string     `mapstructure:"description"`
	Version     string     `mapstructure:"version"`
	LoggerLevel string     `mapstructure:"logger_level"`
	Web         *WebConfig `mapstructure:"web"`
}

// Interface implementation for Application
func (a *Application) GetName() string {
	return a.Name
}

func (a *Application) GetDescription() string {
	return a.Description
}

func (a *Application) GetVersion() string {
	return a.Version
}

func (a *Application) GetLoggerLevel() string {
	return a.LoggerLevel
}

func (a *Application) GetWeb() WebProvider {
	return a.Web
}
