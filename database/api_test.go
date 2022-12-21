package database

import (
	"strings"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPISlowLog(t *testing.T) {
	spec := newSpecification()
	logger := log.NewMockLogger(gomock.NewController(t))

	record := &TestRecord{Name: generator.Name()}

	api := NewAPI(spec.DB, nil, logger)

	logger.EXPECT().Debug(gomock.Any()).Do(func(arg1 interface{}) {
		err := arg1.(error)
		assert.True(t, strings.Contains(err.Error(), "query execution time:"))
	}).AnyTimes()

	err := api.Create(record)
	require.NoError(t, err)

	api.SetSlowQueryDuration(0)

	err = api.Begin()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = api.Rollback()
	})

	logger.EXPECT().Error(gomock.Any()).Do(func(arg1 interface{}) {
		err := arg1.(error)
		assert.True(t, strings.Contains(err.Error(), "query execution time:"))
		assert.True(t, strings.Contains(err.Error(), "query:"))
	}).AnyTimes()
	record = &TestRecord{Name: generator.Name()}

	err = api.Create(record)
	require.NoError(t, err)

	logger.EXPECT().Error(gomock.Any()).Do(func(arg1 interface{}) {
		err := arg1.(error)
		assert.True(t, strings.Contains(err.Error(), "query execution time:"))
		assert.True(t, strings.Contains(err.Error(), "query:"))
	}).AnyTimes()
	record = &TestRecord{Name: generator.Name()}
	err = api.Create(record)
	require.NoError(t, err)
}
