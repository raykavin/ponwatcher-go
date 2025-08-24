package config

// ConfigProvider defines methods for accessing application configuration
type ConfigProvider interface {
	GetApplication() ApplicationProvider
	GetServices() ServicesProvider
}

// ApplicationProvider defines methods for accessing application settings
type ApplicationProvider interface {
	GetName() string
	GetDescription() string
	GetVersion() string
	GetLoggerLevel() string
	GetWeb() WebProvider
}

// ServicesProvider defines methods for accessing service configurations
type ServicesProvider interface {
	GetUNM() UNMProvider
}

// WebProvider defines methods for accessing web server configuration
type WebProvider interface {
	GetListen() int
	GetSSLEnabled() bool
	GetSSLCert() string
	GetSSLKey() string
	GetNoRouteTo() string
	GetCORS() CORSProvider
}

// CORSProvider defines methods for accessing CORS configuration
type CORSProvider interface {
	GetAllowOrigins() []string
	GetAllowMethods() []string
	GetAllowHeaders() []string
}

// UNMProvider defines methods for accessing UNM service configuration
type UNMProvider interface {
	GetHost() string
	GetPort() uint16
	GetUsername() string
	GetPassword() string
}
