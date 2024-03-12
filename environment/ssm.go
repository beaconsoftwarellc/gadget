package environment

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/beaconsoftwarellc/gadget/v2/log"
)

//go:generate mockgen -source=$GOFILE -package environment -destination ssmclient_mock.gen.go
type ssmClient interface {
	GetParametersByPath(*ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error)
}

// SSM wraps the SSM client with an in memory cache
type SSM struct {
	cache       map[string]map[string]string
	client      ssmClient
	environment string
}

// NewSSM returns a SSM for the environment with a client and an initialized cache
func NewSSM(environment string) *SSM {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return &SSM{
		cache:       make(map[string]map[string]string),
		client:      ssm.New(session),
		environment: environment,
	}
}

// Has checks for a given key in the cache
func (s *SSM) Has(path, name string) (string, bool) {
	path = fmt.Sprintf(path, s.environment)
	if val, ok := s.cache[path]; ok {
		value, ok := val[name]
		return value, ok
	}
	return "", false
}

// Add a map of data from SSM to the cache
func (s *SSM) Add(path string, data map[string]string) {
	s.cache[path] = data
}

// Get a value from the cache, if it is not found it will load from SSM
func (s *SSM) Get(path, name string, logger log.Logger) string {
	if value, ok := s.Has(path, name); ok {
		return value
	}
	err := s.loadSSMParameters(path)
	if err != nil {
		logger.Errorf("Issue loading from SSM, %s (%s)", path, err)
		return ""
	}
	if value, ok := s.Has(path, name); ok {
		return value
	}
	return ""
}

func (s *SSM) loadSSMParameters(path string) error {
	path = fmt.Sprintf(path, s.environment)
	results := make(map[string]string)
	params := &ssm.GetParametersByPathInput{
		Path:       &path,
		MaxResults: aws.Int64(10),
	}
	for {
		resp, err := s.client.GetParametersByPath(params)
		if err != nil {
			return err
		}
		for _, p := range resp.Parameters {
			results[*p.Name] = *p.Value
		}
		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}
	s.Add(path, results)
	return nil
}
