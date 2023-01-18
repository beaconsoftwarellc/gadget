package sqs

import "github.com/aws/aws-sdk-go-v2/aws"

// NewEndpointResolver for overriding default endpoint resolution.
// iow don't let aws overwrite the url you have to pass in.
func NewEndpointResolver(disableHTTPS bool) aws.EndpointResolverWithOptions {

}

type endpointResolver struct {
}

func (er *endpointResolver) ResolveEndpoint(service, region string,
	options ...interface{}) (aws.Endpoint, error) {

}
