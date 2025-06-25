package email

type Config struct {
	ValidateForm bool
	ValidateDNS  bool
}

func newConfig() Config {
	return Config{
		ValidateForm: true,
		ValidateDNS:  false,
	}
}

func WithFormValidation(enabled bool) func(Config) Config {
	return func(c Config) Config {
		c.ValidateForm = enabled
		return c
	}
}

func WithDNSValidation(enabled bool) func(Config) Config {
	return func(c Config) Config {
		c.ValidateDNS = enabled
		return c
	}
}
