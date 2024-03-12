package environment

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	assert1 "github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func Test_ssm_Has(t *testing.T) {
	assert := assert1.New(t)

	ssm := NewSSM("env")
	path := "%s-foo"
	name := "bar"

	value, ok := ssm.Has(path, name)
	assert.Empty(value)
	assert.False(ok)

	expected := "baz"
	ssm.Add("env-foo", map[string]string{name: expected})
	value, ok = ssm.Has(path, name)
	assert.Equal(expected, value)
	assert.True(ok)
}

func Test_ssm_loadSSMParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockClient := NewMockssmClient(ctrl)
	ssmClient := &SSM{
		cache:       make(map[string]map[string]string),
		client:      mockClient,
		environment: "env",
	}
	path := "%s-foo"

	t.Run("error", func(t *testing.T) {
		assert := assert1.New(t)
		expected := generator.ID("err")
		mockClient.EXPECT().GetParametersByPath(&ssm.GetParametersByPathInput{
			MaxResults: aws.Int64(10),
			Path:       aws.String("env-foo"),
		}).Return(nil, errors.New(expected))

		err := ssmClient.loadSSMParameters(path)
		assert.EqualError(err, expected)
	})

	t.Run("no results", func(t *testing.T) {
		assert := assert1.New(t)
		mockClient.EXPECT().GetParametersByPath(&ssm.GetParametersByPathInput{
			MaxResults: aws.Int64(10),
			Path:       aws.String("env-foo"),
		}).Return(&ssm.GetParametersByPathOutput{}, nil)

		err := ssmClient.loadSSMParameters(path)
		assert.NoError(err)
	})

	t.Run("single result", func(t *testing.T) {
		assert := assert1.New(t)
		expected := generator.ID("val")
		mockClient.EXPECT().GetParametersByPath(&ssm.GetParametersByPathInput{
			MaxResults: aws.Int64(10),
			Path:       aws.String("env-foo"),
		}).Return(&ssm.GetParametersByPathOutput{
			Parameters: []*ssm.Parameter{
				{
					Name:  aws.String("bar"),
					Value: aws.String(expected),
				},
			},
		}, nil)

		err := ssmClient.loadSSMParameters(path)
		assert.NoError(err)
		value, ok := ssmClient.Has(path, "bar")
		assert.True(ok)
		assert.Equal(expected, value)
	})

	t.Run("calls twice with NextToken", func(t *testing.T) {
		assert := assert1.New(t)
		expected := generator.ID("val")
		mockClient.EXPECT().GetParametersByPath(&ssm.GetParametersByPathInput{
			MaxResults: aws.Int64(10),
			Path:       aws.String("env-foo"),
		}).Return(&ssm.GetParametersByPathOutput{
			NextToken: aws.String("next"),
			Parameters: []*ssm.Parameter{
				{
					Name:  aws.String("bar"),
					Value: aws.String(expected),
				},
			},
		}, nil)
		expected1 := generator.ID("val")
		mockClient.EXPECT().GetParametersByPath(&ssm.GetParametersByPathInput{
			MaxResults: aws.Int64(10),
			Path:       aws.String("env-foo"),
			NextToken:  aws.String("next"),
		}).Return(&ssm.GetParametersByPathOutput{
			Parameters: []*ssm.Parameter{
				{
					Name:  aws.String("baz"),
					Value: aws.String(expected1),
				},
			},
		}, nil)

		err := ssmClient.loadSSMParameters(path)
		assert.NoError(err)
		value, ok := ssmClient.Has(path, "bar")
		assert.True(ok)
		assert.Equal(expected, value)
		value, ok = ssmClient.Has(path, "baz")
		assert.True(ok)
		assert.Equal(expected1, value)
	})

}

func Test_ssm_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockClient := NewMockssmClient(ctrl)
	ssmClient := &SSM{
		cache:       make(map[string]map[string]string),
		client:      mockClient,
		environment: "env",
	}
	path := "%s-foo"
	logger := log.NewStackLogger()

	t.Run("cached", func(t *testing.T) {
		assert := assert1.New(t)
		expected := generator.ID("val")
		ssmClient.Add("env-foo", map[string]string{"bar": expected})

		value := ssmClient.Get(path, "bar", logger)
		assert.Equal(expected, value)
	})

	t.Run("load error", func(t *testing.T) {
		assert := assert1.New(t)
		name := generator.ID("name")
		expected := generator.ID("err")
		mockClient.EXPECT().GetParametersByPath(&ssm.GetParametersByPathInput{
			MaxResults: aws.Int64(10),
			Path:       aws.String("env-foo"),
		}).Return(nil, errors.New(expected))

		actual := ssmClient.Get(path, name, logger)
		assert.Empty(actual)
	})

	t.Run("load success", func(t *testing.T) {
		assert := assert1.New(t)
		name := generator.ID("name")
		expected := generator.ID("val")
		mockClient.EXPECT().GetParametersByPath(&ssm.GetParametersByPathInput{
			MaxResults: aws.Int64(10),
			Path:       aws.String("env-foo"),
		}).Return(&ssm.GetParametersByPathOutput{
			Parameters: []*ssm.Parameter{
				{
					Name:  aws.String(name),
					Value: aws.String(expected),
				},
			},
		}, nil)

		actual := ssmClient.Get(path, name, logger)
		assert.Equal(expected, actual)
	})

	t.Run("dne", func(t *testing.T) {
		assert := assert1.New(t)
		name := generator.ID("name")
		mockClient.EXPECT().GetParametersByPath(&ssm.GetParametersByPathInput{
			MaxResults: aws.Int64(10),
			Path:       aws.String("env-foo"),
		}).Return(&ssm.GetParametersByPathOutput{}, nil)

		actual := ssmClient.Get(path, name, logger)
		assert.Empty(actual)
	})
}
