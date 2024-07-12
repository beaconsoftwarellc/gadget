package environment

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	assert1 "github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func Test_ssm_Has(t *testing.T) {
	assert := assert1.New(t)

	obj := NewSSM(context.Background(), "region", "env", "proj", log.Global())
	ssm := obj.(*ssmAddGet)

	path := "foo"
	name := "bar"
	assert.Equal("/env-proj/", ssm.getPath(""))

	value, ok := ssm.Get(path, name)
	assert.Empty(value)
	assert.False(ok)

	expected := "baz"
	ssm.add("/env-foo/", map[string]interface{}{name: expected})
	value, ok = ssm.Has(path, name)
	assert.Equal(expected, value)
	assert.True(ok)

	ssm.add(ssm.getPath(""), map[string]interface{}{name: expected})
	value, ok = ssm.Has("", name)
	assert.Equal(expected, value)
	assert.True(ok)
}

func Test_ssm_loadSSMParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockClient := NewMockssmClient(ctrl)
	ssmClient := &ssmAddGet{
		cache:          make(map[string]map[string]interface{}),
		client:         mockClient,
		environment:    "env",
		defaultProject: "proj",
	}
	path := "foo"

	t.Run("error", func(t *testing.T) {
		assert := assert1.New(t)
		expected := generator.ID("err")
		mockClient.EXPECT().GetParametersByPath(
			gomock.Any(),
			&ssm.GetParametersByPathInput{
				MaxResults: aws.Int32(10),
				Path:       aws.String("foo"),
			}).Return(nil, errors.New(expected))

		err := ssmClient.loadSSMParameters(path)
		assert.EqualError(err, expected)
	})

	t.Run("no results", func(t *testing.T) {
		assert := assert1.New(t)
		mockClient.EXPECT().GetParametersByPath(gomock.Any(),
			&ssm.GetParametersByPathInput{
				MaxResults: aws.Int32(10),
				Path:       aws.String("foo"),
			}).Return(&ssm.GetParametersByPathOutput{}, nil)

		err := ssmClient.loadSSMParameters(path)
		assert.NoError(err)
	})

	t.Run("single result", func(t *testing.T) {
		assert := assert1.New(t)
		expected := generator.ID("val")
		mockClient.EXPECT().GetParametersByPath(gomock.Any(),
			&ssm.GetParametersByPathInput{
				MaxResults: aws.Int32(10),
				Path:       aws.String("foo"),
			}).Return(&ssm.GetParametersByPathOutput{
			Parameters: []types.Parameter{
				{
					Name:  aws.String("bar"),
					Value: aws.String(expected),
				},
			},
		}, nil)

		err := ssmClient.loadSSMParameters(path)
		assert.NoError(err)
		value, ok := ssmClient.getParameter(path, "bar")
		assert.True(ok)
		assert.Equal(expected, value)
	})

	t.Run("calls twice with NextToken", func(t *testing.T) {
		assert := assert1.New(t)
		expected := generator.ID("val")
		mockClient.EXPECT().GetParametersByPath(gomock.Any(),
			&ssm.GetParametersByPathInput{
				MaxResults: aws.Int32(10),
				Path:       aws.String("foo"),
			}).Return(&ssm.GetParametersByPathOutput{
			NextToken: aws.String("next"),
			Parameters: []types.Parameter{
				{
					Name:  aws.String("bar"),
					Value: aws.String(expected),
				},
			},
		}, nil)
		expected1 := generator.ID("val")
		mockClient.EXPECT().GetParametersByPath(gomock.Any(),
			&ssm.GetParametersByPathInput{
				MaxResults: aws.Int32(10),
				Path:       aws.String("foo"),
				NextToken:  aws.String("next"),
			}).Return(&ssm.GetParametersByPathOutput{
			Parameters: []types.Parameter{
				{
					Name:  aws.String("baz"),
					Value: aws.String(expected1),
				},
			},
		}, nil)

		err := ssmClient.loadSSMParameters(path)
		assert.NoError(err)
		value, ok := ssmClient.getParameter(path, "bar")
		assert.True(ok)
		assert.Equal(expected, value)
		value, ok = ssmClient.getParameter(path, "baz")
		assert.True(ok)
		assert.Equal(expected1, value)
	})
}

func Test_ssm_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockClient := NewMockssmClient(ctrl)
	newSsmClient := func() *ssmAddGet {
		return &ssmAddGet{
			cache:   make(map[string]map[string]interface{}),
			client:  mockClient,
			context: context.Background(),
			logger:  log.NewStackLogger(),
		}
	}
	project := "foo"

	t.Run("cached", func(t *testing.T) {
		ssmClient := newSsmClient()
		assert := assert1.New(t)
		expected := generator.ID("val")
		ssmClient.Add(project,
			map[string]interface{}{"bar": expected})

		value, ok := ssmClient.Get(project, "bar")
		assert.True(ok)
		assert.Equal(expected, value)
	})

	t.Run("load error", func(t *testing.T) {
		ssmClient := newSsmClient()
		assert := assert1.New(t)
		name := generator.ID("name")
		expected := generator.ID("err")
		mockClient.EXPECT().GetParametersByPath(
			gomock.Any(),
			&ssm.GetParametersByPathInput{
				MaxResults: aws.Int32(10),
				Path:       aws.String(ssmClient.getPath(project)),
			}).Return(nil, errors.New(expected))

		actual, ok := ssmClient.Get(project, name)
		assert.False(ok)
		assert.Nil(actual)
	})

	t.Run("load success", func(t *testing.T) {
		ssmClient := newSsmClient()
		assert := assert1.New(t)
		name := generator.ID("name")
		expected := generator.ID("val")
		mockClient.EXPECT().GetParametersByPath(
			gomock.Any(),
			&ssm.GetParametersByPathInput{
				MaxResults: aws.Int32(10),
				Path:       aws.String(ssmClient.getPath(project)),
			}).Return(&ssm.GetParametersByPathOutput{
			Parameters: []types.Parameter{
				{
					Name:  aws.String(name),
					Value: aws.String(expected),
				},
			},
		}, nil)

		actual, ok := ssmClient.Get(project, name)
		assert.True(ok)
		assert.Equal(expected, actual)
	})

	t.Run("dne", func(t *testing.T) {
		ssmClient := newSsmClient()
		assert := assert1.New(t)
		name := generator.ID("name")
		mockClient.EXPECT().GetParametersByPath(
			gomock.Any(),
			&ssm.GetParametersByPathInput{
				MaxResults: aws.Int32(10),
				Path:       aws.String(ssmClient.getPath(project)),
			},
		).Return(&ssm.GetParametersByPathOutput{}, nil)

		actual, ok := ssmClient.Get(project, name)
		assert.False(ok)
		assert.Nil(actual)
	})
}
