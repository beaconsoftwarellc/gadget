package environment

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

// env file name is ${Environment}-${Project}-env.json
const s3ItemFmt = "%s-%s-env.json"

//go:generate mockgen -source=$GOFILE -package environment -destination s3client_mock.gen.go
type s3Client interface {
	Download(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (n int64, err error)
}

// Bucket wraps the S3 downloader with an in memory cache
type Bucket struct {
	cache          map[string]map[string]interface{}
	downloader     s3Client
	bucketName     string
	defaultProject string
	environment    string
}

// NewBucket returns a Bucket with an S3 downloader and an initialized cache
func NewBucket(bucketName, environment, project string) *Bucket {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &Bucket{
		cache:          make(map[string]map[string]interface{}),
		downloader:     s3manager.NewDownloader(sess),
		bucketName:     bucketName,
		defaultProject: project,
		environment:    environment,
	}
}

func (b *Bucket) getItemName(project string) string {
	if stringutil.IsWhiteSpace(project) {
		project = b.defaultProject
	}
	return fmt.Sprintf(s3ItemFmt, b.environment, project)
}

// Has checks for a given key in the cache
func (b *Bucket) Has(project, key string) (interface{}, bool) {
	if val, ok := b.cache[b.getItemName(project)]; ok {
		value, ok := val[key]
		return value, ok
	}
	return nil, false
}

// Add a map of data from S3 to the cache
func (b *Bucket) Add(project string, data map[string]interface{}) {
	b.cache[b.getItemName(project)] = data
}

// Get pulls a value from a map loaded from an S3 bucket
func (b *Bucket) Get(project, key string, logger log.Logger) interface{} {
	if value, ok := b.Has(project, key); ok {
		return value
	}
	data := aws.NewWriteAtBuffer(make([]byte, 0))

	items := make(map[string]interface{})

	s3Item := b.getItemName(project)
	_, err := b.downloader.Download(data,
		&s3.GetObjectInput{
			Bucket: aws.String(b.bucketName),
			Key:    aws.String(s3Item),
		})
	if err != nil {
		logger.Errorf("Issue loading from S3, %s/%s (%s)", b.bucketName, s3Item, err)
		return nil
	}

	err = json.Unmarshal(data.Bytes(), &items)
	if err != nil {
		logger.Errorf("Issue unmarshalling from S3, %s/%s (%s)", b.bucketName, s3Item, err)
		return nil
	}

	b.Add(project, items)

	value, ok := items[key]
	if !ok {
		return nil
	}
	return value
}
