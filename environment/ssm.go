package environment

import (
	"context"
	"fmt"

	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

// default path is /${Environment}-${Project}/
const ssmPathFmt = "/%s-%s/"

//go:generate mockgen -source=$GOFILE -package environment -destination ssmclient_mock.gen.go
type ssmClient interface {
	GetParametersByPath(context.Context, *ssm.GetParametersByPathInput,
		...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error)
}

// ssmAddGet wraps the ssmAddGet client with an in memory cache
type ssmAddGet struct {
	cache          map[string]map[string]interface{}
	client         ssmClient
	defaultProject string
	environment    string
	context        context.Context
	logger         log.Logger
}

// NewSSM returns a SSM for the environment with a client and an initialized cache
func NewSSM(ctx context.Context, region, environment, project string,
	logger log.Logger) AddGet {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		panic(log.Fatalf("[ENV.SSM.42] failed to load default config: %s", err))
	}
	return &ssmAddGet{
		cache:          make(map[string]map[string]interface{}),
		client:         ssm.NewFromConfig(cfg),
		defaultProject: project,
		environment:    environment,
		context:        ctx,
		logger:         logger,
	}
}

func (s *ssmAddGet) getPath(project string) string {
	if stringutil.IsWhiteSpace(project) {
		project = s.defaultProject
	}
	return fmt.Sprintf(ssmPathFmt, s.environment, project)
}

// Has checks for a given key in the cache
func (s *ssmAddGet) Has(project, name string) (interface{}, bool) {
	return s.getParameter(s.getPath(project), name)
}

// Add a map of data from SSM to the cache
func (s *ssmAddGet) Add(project string, data map[string]interface{}) {
	path := s.getPath(project)
	s.add(path, data)
}

func (s *ssmAddGet) add(path string, data map[string]interface{}) {
	s.cache[path] = data
}

func (s *ssmAddGet) getParameter(path, name string) (interface{}, bool) {
	var (
		val   map[string]interface{}
		value interface{}
		ok    bool
	)
	if val, ok = s.cache[path]; ok {
		value, ok = val[name]
	}
	return value, ok
}

// Get a value from the cache, if it is not found it will load from SSM
func (s *ssmAddGet) Get(project, name string) (interface{}, bool) {
	path := s.getPath(project)
	if value, ok := s.getParameter(path, name); ok {
		return value, ok
	}
	err := s.loadSSMParameters(path)
	if err != nil {
		s.logger.Errorf("Issue loading from SSM, %s (%s)", path, err)
		return nil, false
	}
	return s.getParameter(path, name)
}

func (s *ssmAddGet) loadSSMParameters(path string) error {
	results := make(map[string]interface{})
	params := &ssm.GetParametersByPathInput{
		Path:       &path,
		MaxResults: aws.Int32(10),
	}
	for {
		resp, err := s.client.GetParametersByPath(s.context, params)
		if err != nil {
			return err
		}
		for _, p := range resp.Parameters {
			results[strings.TrimPrefix(*p.Name, path)] = *p.Value
		}
		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}
	s.add(path, results)
	return nil
}
