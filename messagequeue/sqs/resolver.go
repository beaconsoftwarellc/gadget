package sqs

import (
	"context"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

// NewStaticEndpointResolverV2 returns a new type resolver that
// always returns the same endpoint containing the passed host.
func NewStaticEndpointResolverV2(host string) sqs.EndpointResolverV2 {
	return &staticResolver{
		smithyendpoints.Endpoint{
			URI:     url.URL{Scheme: "http", Host: host},
			Headers: http.Header{},
		},
	}
}

// StaticEndpointResolverOption returns an option that configures the
// sqs client to use the passed host.
func StaticEndpointResolverOption(host string) func(*sqs.Options) {
	return func(o *sqs.Options) {
		o.EndpointResolverV2 = NewStaticEndpointResolverV2(host)
	}
}

type staticResolver struct {
	endpoint smithyendpoints.Endpoint
}

func (resolver *staticResolver) ResolveEndpoint(context.Context, sqs.EndpointParameters) (
	smithyendpoints.Endpoint, error,
) {
	return resolver.endpoint, nil
}
