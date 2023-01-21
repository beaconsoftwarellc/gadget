package sqs

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	assert1 "github.com/stretchr/testify/assert"
)

func TestNewEndpointResolver(t *testing.T) {
	assert := assert1.New(t)
	actual := NewEndpointResolver("", "")
	assert.NotNil(actual)
}

func TestEndpointResolverResolveEndpoint(t *testing.T) {
	assert := assert1.New(t)
	region := generator.String(32)
	expected := generator.String(32)
	resolver := NewEndpointResolver(region, expected)
	// wrong service
	_, err := resolver.ResolveEndpoint(
		generator.String(5), region)
	assert.EqualError(err, (&aws.EndpointNotFoundError{}).Error())
	// correct service
	actual, err := resolver.ResolveEndpoint(ServiceName, region)
	assert.NoError(err)
	assert.Equal(expected, actual.URL)
	assert.Equal(true, actual.HostnameImmutable)
}

func TestEndpointResolverResolveEndpoint_NotFound(t *testing.T) {
	assert := assert1.New(t)
	resolver := NewEndpointResolver(generator.String(32), generator.String(23))
	_, actual := resolver.ResolveEndpoint(
		ServiceName, generator.String(5))
	expected := (&aws.EndpointNotFoundError{}).Error()
	assert.EqualError(actual, expected)
}
