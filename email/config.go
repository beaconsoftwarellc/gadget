package email

type Config struct {
	ValidateForm bool
}

func newConfig() Config {
	return Config{
		ValidateForm: true,
	}
}

func WithFormValidation(enabled bool) func(Config) Config {
	return func(c Config) Config {
		c.ValidateForm = enabled
		return c
	}
}
