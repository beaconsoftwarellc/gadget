package environment

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/beaconsoftwarellc/gadget/log"
)

// Bucket wraps the S3 downloader with an in memory cache
type Bucket struct {
	cache      map[string]map[string]interface{}
	Downloader *s3manager.Downloader
}

// NewBucket returns a Bucket with an S3 downloader and an initialized cache
func NewBucket() *Bucket {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &Bucket{
		cache:      make(map[string]map[string]interface{}),
		Downloader: s3manager.NewDownloader(sess),
	}
}

// Has checks for a given key in the cache
func (b *Bucket) Has(bucket, item, key string) (interface{}, bool) {
	if val, ok := b.cache[bucket+item]; ok {
		value, ok := val[key]
		return value, ok
	}
	return nil, false
}

// Add a map of data from S3 to the cache
func (b *Bucket) Add(bucket, item string, data map[string]interface{}) {
	b.cache[bucket+item] = data
}

// Get pulls a value from a map loaded from and S3 bucket
func (b *Bucket) Get(bucket, item, key string) interface{} {
	if value, ok := b.Has(bucket, item, key); ok {
		return value
	}
	data := aws.NewWriteAtBuffer(make([]byte, 0))

	items := make(map[string]interface{})

	_, err := b.Downloader.Download(data,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		log.Errorf("Issue loading from S3, %s/%s (%s)", bucket, item, err)
		return nil
	}

	err = json.Unmarshal(data.Bytes(), &items)
	if err != nil {
		log.Errorf("Issue unmarshalling from S3, %s/%s (%s)", bucket, item, err)
		return nil
	}

	b.Add(bucket, item, items)

	value, ok := items[key]
	if !ok {
		return nil
	}
	return value
}
