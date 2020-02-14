package environment

import (
	"io/ioutil"
	"net/http"
	"strings"

	"gitlab.com/beacon-software/gadget/log"
)

const awsMetaService = "aws-meta://"

func isAWSMetaServiceLookup(host string) bool {
	return strings.HasPrefix(host, awsMetaService)
}

// AWSLookup tries to perform a Get against the provided URL
// This is a work around intended to get the Local Host IP / Name for the EC2 instance that service is running on
// through the AWS Meta data service since we can't dynamically set that environment variable.
func AWSLookup(host string) string {
	target := strings.TrimLeft(host, awsMetaService)
	resp, err := http.Get("http://" + target)
	if err != nil {
		log.Errorf("Unable to performs AWS Lookup against %s", target)
		return host
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Read from %s failed with %s", target, err)
		return host
	}
	log.Infof("AWS Lookup: %s", string(body))
	return string(body)
}
