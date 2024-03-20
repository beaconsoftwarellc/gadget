package environment

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

// default path is /${Environment}-${Project}/
const ssmPathFmt = "/%s-%s/"

//go:generate mockgen -source=$GOFILE -package environment -destination ssmclient_mock.gen.go
type ssmClient interface {
	GetParametersByPath(*ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error)
}

// SSM wraps the SSM client with an in memory cache
type SSM struct {
	cache       map[string]map[string]string
	client      ssmClient
	defaultPath string
}

// NewSSM returns a SSM for the environment with a client and an initialized cache
func NewSSM(environment, project string) *SSM {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return &SSM{
		cache:       make(map[string]map[string]string),
		client:      ssm.New(session),
		defaultPath: fmt.Sprintf(ssmPathFmt, environment, project),
	}
}

// Has checks for a given key in the cache
func (s *SSM) Has(path, name string) (string, bool) {
	if stringutil.IsWhiteSpace(path) {
		path = s.defaultPath
	}
	return s.getParameter(path, name)
}

func (s *SSM) getParameter(path, name string) (string, bool) {
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
	if stringutil.IsWhiteSpace(path) {
		path = s.defaultPath
	}
	if value, ok := s.cache[path]; ok {
		return value[name]
	}
	err := s.loadSSMParameters(path)
	if err != nil {
		logger.Errorf("Issue loading from SSM, %s (%s)", path, err)
		return ""
	}
	value, _ := s.getParameter(path, name)
	return value
}

func (s *SSM) loadSSMParameters(path string) error {
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

func (s *SSM) clearCache() {
	s.cache = make(map[string]map[string]string)
}
