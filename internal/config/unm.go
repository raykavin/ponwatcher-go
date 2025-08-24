package config

// UNMConfig contains configuration for UNM service connection
type UNMConfig struct {
	Host     string `mapstructure:"host"`
	Port     uint16 `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// Interface implementation for UNMConfig
func (u *UNMConfig) GetHost() string {
	return u.Host
}

func (u *UNMConfig) GetPort() uint16 {
	return u.Port
}

func (u *UNMConfig) GetUsername() string {
	return u.Username
}

func (u *UNMConfig) GetPassword() string {
	return u.Password
}
