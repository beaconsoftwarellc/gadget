package environment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

// env file name is ${Environment}-${Project}-env.json
const s3ItemFmt = "%s-%s-env.json"

//go:generate mockgen -source=$GOFILE -package environment -destination s3_mock.gen.go

type s3Client interface {
	Download(ctx context.Context, w io.WriterAt, input *s3.GetObjectInput,
		options ...func(*manager.Downloader)) (n int64, err error)
}

// awsbucket wraps the S3 downloader with an in memory cache
type awsbucket struct {
	cache          map[string]map[string]interface{}
	downloader     s3Client
	bucketName     string
	defaultProject string
	environment    string
	context        context.Context
	logger         log.Logger
}

// NewBucket returns a Bucket with an S3 downloader and an initialized cache
func NewBucket(ctx context.Context, region, bucketName, environment,
	project string, logger log.Logger) AddGet {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		panic(log.Fatalf("[ENV.S3.41] failed to load default config: %s", err))
	}
	return &awsbucket{
		context:        ctx,
		cache:          make(map[string]map[string]interface{}),
		downloader:     manager.NewDownloader(s3.NewFromConfig(cfg)),
		bucketName:     bucketName,
		defaultProject: project,
		environment:    environment,
		logger:         logger,
	}
}

func (b *awsbucket) getItemName(project string) string {
	if stringutil.IsWhiteSpace(project) {
		project = b.defaultProject
	}
	return fmt.Sprintf(s3ItemFmt, b.environment, project)
}

// Add a map of data from S3 to the cache
func (b *awsbucket) Add(project string, data map[string]interface{}) {
	b.cache[b.getItemName(project)] = data
}

// Get pulls a value from a map loaded from an S3 bucket
func (b *awsbucket) Get(project, key string) (interface{}, bool) {
	if val, ok := b.cache[b.getItemName(project)]; ok {
		value, ok := val[key]
		return value, ok
	}
	data := manager.NewWriteAtBuffer(make([]byte, 0))

	items := make(map[string]interface{})

	s3Item := b.getItemName(project)
	_, err := b.downloader.Download(b.context, data,
		&s3.GetObjectInput{
			Bucket: aws.String(b.bucketName),
			Key:    aws.String(s3Item),
		})
	if err != nil {
		b.logger.Errorf("Issue loading from S3, %s/%s (%s)",
			b.bucketName, s3Item, err)
		return nil, false
	}

	err = json.Unmarshal(data.Bytes(), &items)
	if err != nil {
		b.logger.Errorf("Issue unmarshalling from S3, %s/%s (%s)",
			b.bucketName, s3Item, err)
		return nil, false
	}

	b.Add(project, items)

	value, ok := items[key]
	if !ok {
		return nil, false
	}
	return value, true
}
