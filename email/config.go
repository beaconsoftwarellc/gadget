package email

type Config struct {
	ValidateForm       bool
	ValidateDNS        bool
	ValidateDisposable bool
}

func newConfig() Config {
	return Config{
		ValidateForm:       true,
		ValidateDNS:        false,
		ValidateDisposable: false,
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

func WithDisposableValidation(enabled bool) func(Config) Config {
	return func(c Config) Config {
		c.ValidateDisposable = enabled
		return c
	}
}
