package kinesis

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

type client struct {
	shard   string
	kinesis kinesis.Client
}

func (c *client) Initialize(ctx context.Context) error {
	// we have to initialize the shard if it is empty
	if !stringutil.IsWhiteSpace(c.shard) {
		return nil
	}
	// check our coordination (redis) for a sharditerator, if there is not
	// one we will need to just get the default
	c.kinesis.GetShardIterator(ctx, &kinesis.GetShardIteratorInput{})
}

func (c *client) Something(ctx context.Context) {
	output, err := c.kinesis.GetRecords(ctx, &kinesis.GetRecordsInput{
		ShardIterator: &c.shard,
	})
	output.Records
	c.kinesis.PutRecord(ctx, &kinesis.PutRecordInput{})
}
