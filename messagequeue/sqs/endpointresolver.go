package sqs

import "github.com/aws/aws-sdk-go-v2/aws"

const (
	// ServiceName as expected by the endpoint resolver for SQS
	ServiceName = "SQS"
)

// NewEndpointResolver for overriding default endpoint resolution.
// iow don't let aws overwrite the url you have to pass in.
func NewEndpointResolver(region, queueURL string) aws.EndpointResolverWithOptions {
	resolver := &endpointResolver{
		resolution: make(map[string]string),
	}
	resolver.resolution[region] = queueURL
	return resolver
}

type endpointResolver struct {
	resolution map[string]string
}

func (er *endpointResolver) ResolveEndpoint(service,
	region string, options ...interface{}) (aws.Endpoint, error) {
	var (
		endpoint aws.Endpoint
		err      error
		ok       bool
	)
	if service != ServiceName {
		return endpoint, &aws.EndpointNotFoundError{}
	}
	endpoint.URL, ok = er.resolution[region]
	if ok {
		endpoint.HostnameImmutable = true
	} else {
		err = &aws.EndpointNotFoundError{}
	}
	return endpoint, err
}
