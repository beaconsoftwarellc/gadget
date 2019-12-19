package cloudwatch

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/beaconsoftwarellc/gadget/timeutil"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/beaconsoftwarellc/gadget/dispatcher"
	"github.com/beaconsoftwarellc/gadget/errors"
	"github.com/beaconsoftwarellc/gadget/log"
	"github.com/beaconsoftwarellc/gadget/stringutil"
)

// 1 mebibyte is the actual max, but pad with a tenth so we don't have to be
// exact when calculating message size (int(1048576 * 0.9))
const (
	defaultSendWait         = 30 * time.Second
	maxPayloadSizeBytes int = 943718
)

func newSession() (*session.Session, errors.TracerError) {
	session, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	return session, errors.Wrap(err)
}

type administration struct {
	sync.Mutex
	sendWait   time.Duration
	dispatcher dispatcher.Dispatcher
	cwlogs     *cloudwatchlogs.CloudWatchLogs
	logGroups  map[string]*cloudwatchlogs.LogGroup
	logStreams map[string]*cloudwatchlogs.LogStream
	// wrap destinations
	outputs map[string]*output
}

// we only need one of these lazy initialized
var admin = &administration{
	logGroups:  make(map[string]*cloudwatchlogs.LogGroup),
	logStreams: make(map[string]*cloudwatchlogs.LogStream),
	outputs:    make(map[string]*output),
}

// Administration provides a layer that manages the control of cloud watch logs to behave
// like a standard log output.
type Administration interface {
	// GetOutput for the specified group and output name.
	GetOutput(groupName, outputName string, logLevel log.LevelFlag) (log.Output, errors.TracerError)
}

// GetAdministration for cloud watch logs
func GetAdministration() (Administration, errors.TracerError) {
	if nil != admin.cwlogs {
		return admin, nil
	}
	session, err := newSession()
	if nil != err {
		log.Error(err)
		return nil, err
	}
	admin.cwlogs = cloudwatchlogs.New(session)
	err = admin.UpdateLogGroups()
	log.Error(err)
	return admin, errors.Wrap(err)
}

func (cwa *administration) Run() {
	cwa.Lock()
	defer cwa.Unlock()
	timeutil.RunEvery(func() {
		for _, output := range cwa.outputs {
			output.SendEvents()
		}
	}, cwa.sendWait)
}

func createStreamKey(groupName, streamName string) string {
	groupName = EnsureGroupNameIsValid(groupName)
	streamName = EnsureStreamNameIsValid(streamName)
	return fmt.Sprintf("%s.%s", groupName, streamName)
}

// UpdateLogGroups pulls all the existing log groups from CloudWatch and adds
// them to this instance so that they might be used.
// NOTE: We should not have a ton of log groups so holding all of them in memory
// should not be a big deal. The standard maximum number of log groups in AWS
// is 5000.
func (cwa *administration) UpdateLogGroups() errors.TracerError {
	cwa.Lock()
	defer cwa.Unlock()
	var nextToken string
	var err error
	var input *cloudwatchlogs.DescribeLogGroupsInput
	var output *cloudwatchlogs.DescribeLogGroupsOutput

	var limit int64 = 50
	for {
		if stringutil.IsWhiteSpace(nextToken) {
			input = &cloudwatchlogs.DescribeLogGroupsInput{
				Limit: &limit,
			}
		} else {
			input = &cloudwatchlogs.DescribeLogGroupsInput{
				Limit:     &limit,
				NextToken: &nextToken,
			}
		}
		output, err = cwa.cwlogs.DescribeLogGroups(input)
		if nil != err {
			break
		}
		for _, group := range output.LogGroups {
			cwa.logGroups[*group.LogGroupName] = group
		}
		if len(output.LogGroups) < int(limit) || nil == output.NextToken || stringutil.IsWhiteSpace(*output.NextToken) {
			break
		}
		nextToken = *output.NextToken
	}
	return errors.Wrap(err)
}

func (cwa *administration) GetOutput(groupName, streamName string, logLevel log.LevelFlag) (log.Output, errors.TracerError) {
	var err error
	// get the log group
	group, err := cwa.GetLogGroup(groupName)
	if nil != err {
		return nil, errors.Wrap(err)
	}
	// now for the stream
	streamName = EnsureStreamNameIsValid(streamName)
	outputKey := createStreamKey(*group.LogGroupName, streamName)
	logOutput, ok := cwa.outputs[outputKey]
	if !ok {
		stream, err := cwa.GetLogStream(group, streamName)
		if nil != err {
			return nil, errors.Wrap(err)
		}
		// we are gtg
		logOutput = &output{
			name:     createStreamKey(*group.LogGroupName, *stream.LogStreamName),
			group:    group,
			stream:   stream,
			logLevel: logLevel,
			admin:    cwa,
			buffer:   NewEventQueue(),
		}
	}
	return logOutput, errors.Wrap(err)
}

func (cwa *administration) GetLogGroup(groupName string) (*cloudwatchlogs.LogGroup, errors.TracerError) {
	groupName = EnsureGroupNameIsValid(groupName)
	var err error
	cwa.Lock()
	group, ok := cwa.logGroups[groupName]
	cwa.Unlock()
	if ok {
		return group, nil
	}
	// it does not exist as far as we can tell so try creation
	input := &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: &groupName,
		// we can put tags here as well as needed
	}
	// the response from this is a marker so we do not need it.
	_, err = cwa.cwlogs.CreateLogGroup(input)
	if nil != err {
		// error handling, return error unless it is an 'already exists' which means we just
		// didn't know about it yet
		if err, ok := err.(awserr.Error); !ok || err.Code() != cloudwatchlogs.ErrCodeResourceAlreadyExistsException {
			return nil, errors.Wrap(err)
		}
	}
	// update to bring it into the fold
	err = cwa.UpdateLogGroups()
	if nil != err {
		return nil, errors.Wrap(err)
	}
	cwa.Lock()
	group, ok = cwa.logGroups[groupName]
	cwa.Unlock()
	if !ok {
		return nil, errors.New("could not create or find cloud watch logs log group %s", groupName)
	}
	// if creation fails as existing try an update
	return group, errors.Wrap(err)
}

func (cwa *administration) GetLogStream(group *cloudwatchlogs.LogGroup, streamName string) (*cloudwatchlogs.LogStream, errors.TracerError) {
	streamName = EnsureStreamNameIsValid(streamName)
	var err error
	streamKey := createStreamKey(*group.LogGroupName, streamName)
	cwa.Lock()
	stream, ok := cwa.logStreams[streamKey]
	cwa.Unlock()
	if ok {
		return stream, nil
	}
	input := &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  group.LogGroupName,
		LogStreamName: &streamName,
	}
	// return is a marker value which can be ignored.
	_, err = cwa.cwlogs.CreateLogStream(input)
	if nil != err {
		if err, ok := err.(awserr.Error); !ok || err.Code() != cloudwatchlogs.ErrCodeResourceAlreadyExistsException {
			return nil, errors.Wrap(err)
		}
	}
	// now actually get the damn thing
	stream, err = cwa.FindLogStream(*group.LogGroupName, streamName)
	if nil != err {
		return nil, errors.Wrap(err)
	}
	// add the reference to our map
	cwa.Lock()
	cwa.logStreams[streamKey] = stream
	cwa.Unlock()
	return stream, errors.Wrap(err)
}

func (cwa *administration) UpdateLogStream(groupName, streamName string) {
	streamKey := createStreamKey(groupName, streamName)
	stream, err := cwa.FindLogStream(groupName, streamName)
	if nil != err {
		log.Errorf("failed to update log stream: %s", err)
	}
	cwa.Lock()
	s, ok := cwa.logStreams[streamKey]
	if ok {
		// do not replace or existing tasks will lose their reference.
		s.SetUploadSequenceToken(*stream.UploadSequenceToken)
	} else {
		// this would be weird, but handle it just in case
		cwa.logStreams[streamKey] = stream
	}
	cwa.Unlock()
}

func (cwa *administration) FindLogStream(groupName, streamName string) (*cloudwatchlogs.LogStream, errors.TracerError) {
	groupName = EnsureGroupNameIsValid(groupName)
	streamName = EnsureStreamNameIsValid(streamName)
	input := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        &groupName,
		LogStreamNamePrefix: &streamName,
	}
	output, err := cwa.cwlogs.DescribeLogStreams(input)
	if nil != err {
		return nil, errors.Wrap(err)
	}
	for _, stream := range output.LogStreams {
		if *stream.LogStreamName == streamName {
			// we are not locking so don't update anything here
			return stream, nil
		}
	}
	return nil, errors.New("failed to locate log stream '%s' in log group '%s'", streamName, groupName)
}

// output is a wrapper for a log stream that we can attach our interface methods to.
type output struct {
	sync.Mutex
	name     string
	admin    *administration
	logLevel log.LevelFlag
	group    *cloudwatchlogs.LogGroup
	stream   *cloudwatchlogs.LogStream
	buffer   EventQueue
	// token is unique to the stream and must be set to sequence the events correctly
	token *string
}

func (o *output) Level() log.LevelFlag {
	return o.logLevel
}

func (o *output) Log(message log.Message) {
	o.Lock()
	defer o.Unlock()
	payload := message.JSONString()
	// they want milliseconds since epoch and our timestamps are in seconds.
	ts := message.TimestampUnix * 1000
	o.buffer.Push(&cloudwatchlogs.InputLogEvent{
		Message:   &payload,
		Timestamp: &ts,
	})
}

func (o *output) SendEvents() error {
	o.Lock()
	defer o.Unlock()
	if o.buffer.Size() == 0 {
		return nil
	}
	events := make([]*cloudwatchlogs.InputLogEvent, 0)
	sizeBytes := 0
	for o.buffer.Size() > 0 {
		event, err := o.buffer.Peek()
		if nil != err {
			break
		}
		if sizeBytes+len([]byte(*event.Message)) > maxPayloadSizeBytes {
			break
		}
		// actually pop it now
		o.buffer.Pop()
		events = append(events, event)
	}
	if len(events) == 0 {
		return nil
	}
	input := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  o.group.LogGroupName,
		LogStreamName: o.stream.LogStreamName,
		SequenceToken: o.stream.UploadSequenceToken,
	}
	resp, err := o.admin.cwlogs.PutLogEvents(input)
	if err == nil {
		o.stream = o.stream.SetUploadSequenceToken(*resp.NextSequenceToken)
	}
	if awsErr, ok := err.(awserr.Error); !ok || awsErr.Code() != cloudwatchlogs.ErrCodeInvalidSequenceTokenException {
		// our sequence token got out of data so refresh it
		o.admin.UpdateLogStream(*o.group.LogGroupName, *o.stream.LogStreamName)
		// don't log this error
		err = nil
	} else if nil != err {
		log.Error(err)
	}
	// if we still have events run again
	if o.buffer.Size() > 0 && nil == err {
		return o.SendEvents()
	}
	return err
}

var groupNameRegex = regexp.MustCompile("[^a-zA-Z0-9_\\-/.]+")

// EnsureGroupNameIsValid based upon the rules from aws:
// * Log group names must be unique within a region for an AWS account.
// * Log group names can be between 1 and 512 characters long.
// * Log group names consist of the following characters: a-z, A-Z, 0-9,
// 		'_' (underscore), '-' (hyphen), '/' (forward slash), and '.' (period).
func EnsureGroupNameIsValid(name string) string {
	validName := groupNameRegex.ReplaceAllString(name, "")
	if stringutil.IsWhiteSpace(validName) {
		validName = "EmptyLogGroupName"
	}
	return stringutil.SafeSubstring(validName, 0, 511)
}

// EnsureStreamNameIsValid based upon the provided rules from AWS
//	* Log stream names must be unique within the log group.
//	* Log stream names can be between 1 and 512 characters long.
//	* The ':' (colon) and '*' (asterisk) characters are not allowed.
func EnsureStreamNameIsValid(name string) string {
	validName := strings.Replace(name, ":", "", -1)
	validName = strings.Replace(validName, "*", "", -1)
	if stringutil.IsWhiteSpace(validName) {
		validName = "EmptyLogStreamName"
	}
	return stringutil.SafeSubstring(validName, 0, 511)
}
