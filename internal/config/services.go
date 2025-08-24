package config

// Services contains configuration for all external services
type Services struct {
	UNM *UNMConfig `mapstructure:"unm"`
}

func (s *Services) GetUNM() UNMProvider {
	return s.UNM
}
