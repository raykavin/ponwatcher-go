package config

// WebConfig contains HTTP server and CORS configuration
type WebConfig struct {
	// Server settings
	Listen     int    `mapstructure:"listen"`
	SSLEnabled bool   `mapstructure:"ssl_enabled"`
	SSLCert    string `mapstructure:"ssl_cert"`
	SSLKey     string `mapstructure:"ssl_key"`

	// Routing
	NoRouteTo string `mapstructure:"no_route_to"`

	// CORS settings
	CORS CORSConfig `mapstructure:",squash"`
}

// Interface implementation for WebConfig
func (w *WebConfig) GetListen() int {
	return w.Listen
}

func (w *WebConfig) GetSSLEnabled() bool {
	return w.SSLEnabled
}

func (w *WebConfig) GetSSLCert() string {
	return w.SSLCert
}

func (w *WebConfig) GetSSLKey() string {
	return w.SSLKey
}

func (w *WebConfig) GetNoRouteTo() string {
	return w.NoRouteTo
}

func (w *WebConfig) GetCORS() CORSProvider {
	return &w.CORS
}

// CORSConfig contains Cross-Origin Resource Sharing configuration
type CORSConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
	AllowMethods []string `mapstructure:"allow_methods"`
	AllowHeaders []string `mapstructure:"allow_headers"`
}

// Interface implementation for CORSConfig
func (c *CORSConfig) GetAllowOrigins() []string {
	return c.AllowOrigins
}

func (c *CORSConfig) GetAllowMethods() []string {
	return c.AllowMethods
}

func (c *CORSConfig) GetAllowHeaders() []string {
	return c.AllowHeaders
}
