package timeutil

import (
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/beaconsoftwarellc/gadget/log"
)

// TimeToTimestamp returns a protobuf Timestamp from a Time object
func TimeToTimestamp(t time.Time) *timestamp.Timestamp {
	ts, err := ptypes.TimestampProto(t)
	if nil != err {
		log.Errorf("Time to Timestamp error: %s", err.Error())
	}
	return ts
}

// TimestampToTime returns a Time object from a protobuf Timestamp
func TimestampToTime(ts *timestamp.Timestamp) time.Time {
	if nil == ts {
		return time.Time{}
	}
	t, err := ptypes.Timestamp(ts)
	if nil != err {
		log.Errorf("Timestamp to Times error: %s", err.Error())
	}
	return t
}

// TimestampToNullTime returns a mysql.NullTime from a protobuf Timestamp
func TimestampToNullTime(ts *timestamp.Timestamp) mysql.NullTime {
	return TimeToNullTime(TimestampToTime(ts))
}

// TimestampToNilOrTime returns a time.Time or nil (for JSON) from a protobuf Timestamp
func TimestampToNilOrTime(ts *timestamp.Timestamp) *time.Time {
	nt := TimestampToNullTime(ts)
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// NullTimeToTimestamp returns a protobuf Timestamp from a mysql.NullTime
func NullTimeToTimestamp(nt mysql.NullTime) *timestamp.Timestamp {
	var ts *timestamp.Timestamp
	var err error

	if nt.Valid {
		ts, err = ptypes.TimestampProto(nt.Time)
	} else {
		ts, err = ptypes.TimestampProto(time.Time{})
	}
	if nil != err {
		log.Error(err)
	}
	return ts
}

// TimeToNullTime coverts a Time into a NullTime
func TimeToNullTime(t time.Time) mysql.NullTime {
	if t.IsZero() {
		return mysql.NullTime{}
	}
	return mysql.NullTime{
		Time:  t,
		Valid: true,
	}
}
