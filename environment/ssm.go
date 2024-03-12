package environment

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/beaconsoftwarellc/gadget/v2/log"
)

// SSM wraps the SSM client with an in memory cache
type SSM struct {
	cache       map[string]string
	client      *ssm.SSM
	environment string
}

// NewSSM returns a SSM for the environment with a client and an initialized cache
func NewSSM(environment string) *SSM {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return &SSM{
		cache:       make(map[string]string),
		client:      ssm.New(session),
		environment: environment,
	}
}

// Has checks for a given key in the cache
func (s *SSM) Has(name string) (string, bool) {
	key := fmt.Sprintf(name, s.environment)
	value, ok := s.cache[key]
	return value, ok
}

// Add a map of data from SSM to the cache
func (s *SSM) Add(key string, data string) {
	s.cache[key] = data
}

// Get pulls a value the cache, if it is not found it will pull from SSM
func (s *SSM) Get(name string, logger log.Logger) string {
	if value, ok := s.Has(name); ok {
		return value
	}
	key := fmt.Sprintf(name, s.environment)
	params := &ssm.GetParameterInput{
		Name: &key,
	}
	resp, err := s.client.GetParameter(params)
	if err != nil {
		logger.Errorf("Issue loading from SSM, %s (%s)", key, err)
		return ""
	}
	data := *resp.Parameter.Value
	s.Add(key, data)
	return data
}
