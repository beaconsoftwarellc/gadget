package environment

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

const (
	// default bucket is ${Environment}-${Project}-env
	envBucketFmt = "%s-%s-env"
	s3Item       = "env.json"
)

// Bucket wraps the S3 downloader with an in memory cache
type Bucket struct {
	cache          map[string]map[string]interface{}
	Downloader     *s3manager.Downloader
	defaultProject string
	environment    string
}

// NewBucket returns a Bucket with an S3 downloader and an initialized cache
func NewBucket(environment, project string) *Bucket {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &Bucket{
		cache:          make(map[string]map[string]interface{}),
		Downloader:     s3manager.NewDownloader(sess),
		defaultProject: project,
		environment:    environment,
	}
}

func (b *Bucket) getBucket(project string) string {
	if stringutil.IsWhiteSpace(project) {
		project = b.defaultProject
	}
	return fmt.Sprintf(envBucketFmt, b.environment, project)
}

// Has checks for a given key in the cache
func (b *Bucket) Has(project, key string) (interface{}, bool) {
	if val, ok := b.cache[b.getBucket(project)]; ok {
		value, ok := val[key]
		return value, ok
	}
	return nil, false
}

// Add a map of data from S3 to the cache
func (b *Bucket) Add(project string, data map[string]interface{}) {
	b.cache[b.getBucket(project)] = data
}

// Get pulls a value from a map loaded from an S3 bucket
func (b *Bucket) Get(project, key string, logger log.Logger) interface{} {
	if value, ok := b.Has(project, key); ok {
		return value
	}
	data := aws.NewWriteAtBuffer(make([]byte, 0))

	items := make(map[string]interface{})

	bucket := b.getBucket(project)
	_, err := b.Downloader.Download(data,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(s3Item),
		})
	if err != nil {
		logger.Errorf("Issue loading from S3, %s/%s (%s)", bucket, s3Item, err)
		return nil
	}

	err = json.Unmarshal(data.Bytes(), &items)
	if err != nil {
		logger.Errorf("Issue unmarshalling from S3, %s/%s (%s)", bucket, s3Item, err)
		return nil
	}

	b.Add(project, items)

	value, ok := items[key]
	if !ok {
		return nil
	}
	return value
}
