package timeutil

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	assert1 "github.com/stretchr/testify/assert"
)

func TestTimeToTimestamp(t *testing.T) {
	assert := assert1.New(t)
	tm := time.Now()
	ts := TimeToTimestamp(tm)
	assert.Equal(tm.Unix(), ts.Seconds)

	// test uninitialized time
	tm = time.Time{}
	ts = TimeToTimestamp(tm)
	assert.Equal(int64(0), ts.Seconds)
}

func TestTimestampToTime(t *testing.T) {
	assert := assert1.New(t)
	ts, err := ptypes.TimestampProto(time.Now())
	assert.NoError(err)

	tm := TimestampToTime(ts)
	assert.Equal(tm.Unix(), ts.Seconds)

	tm = TimestampToTime(nil)
	assert.True(tm.IsZero())

	tm = TimestampToTime(&timestamp.Timestamp{})
	assert.Equal(int64(0), tm.Unix())
}

func TestTimestampToNullTime(t *testing.T) {
	assert := assert1.New(t)
	ts, err := ptypes.TimestampProto(time.Time{})
	nt := TimestampToNullTime(ts)
	assert.NoError(err)
	assert.False(nt.Valid)

	now := time.Now()
	ts, err = ptypes.TimestampProto(now)
	nt = TimestampToNullTime(ts)
	assert.NoError(err)
	assert.True(nt.Valid)
	assert.Equal(now.Second(), nt.Time.Second())
}

func TestNullTimeToTimestamp(t *testing.T) {
	assert := assert1.New(t)
	nt := mysql.NullTime{}
	ts := NullTimeToTimestamp(nt)
	assert.False(nt.Valid)
	assert.Equal(nt.Time.Unix(), ts.Seconds)

	tm := TimestampToTime(ts)
	assert.True(tm.IsZero())

	nt.Time = time.Now()
	nt.Valid = true
	ts = NullTimeToTimestamp(nt)
	assert.Equal(nt.Time.Unix(), ts.Seconds)

	tm = TimestampToTime(ts)
	assert.False(tm.IsZero())
}

func TestTimestampToNilOrTime(t *testing.T) {
	type args struct {
		ts *timestamp.Timestamp
	}
	nullTime, _ := ptypes.TimestampProto(time.Time{})
	now := time.Now().UTC()
	nowTime, _ := ptypes.TimestampProto(now)
	tests := []struct {
		name string
		args args
		want *time.Time
	}{
		{
			name: "nil",
			args: args{
				ts: nullTime,
			},
			want: nil,
		},
		{
			name: "valid",
			args: args{
				ts: nowTime,
			},
			want: &now,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimestampToNilOrTime(tt.args.ts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TimestampToNilOrTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
