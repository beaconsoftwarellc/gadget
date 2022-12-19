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
	var (
		e0, e1, e2 error
	)
	spec := newSpecification()
	logger := log.NewMockLogger(gomock.NewController(t))

	record := &TestRecord{Name: generator.Name()}

	api := NewAPI(spec.DB, nil, logger)

	logger.EXPECT().Debug(gomock.Any()).Do(func(arg1 interface{}) {
		e0 = arg1.(error)
		assert.True(t, strings.HasPrefix(e0.Error(), "query execution time:"))
	})

	err := api.Create(record)
	require.NoError(t, err)

	api.SetSlowQueryDuration(0)

	err = api.Begin()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = api.Rollback()
	})

	logger.EXPECT().Error(gomock.Any()).Do(func(arg1 interface{}) {
		e1 = arg1.(error)
		assert.True(t, strings.HasPrefix(e1.Error(), "query execution time:"))
	})
	record = &TestRecord{Name: generator.Name()}

	err = api.Create(record)
	require.NoError(t, err)

	logger.EXPECT().Error(gomock.Any()).Do(func(arg1 interface{}) {
		e2 = arg1.(error)
		assert.True(t, strings.HasPrefix(e2.Error(), "query execution time:"))
	})
	record = &TestRecord{Name: generator.Name()}
	err = api.Create(record)
	require.NoError(t, err)

	e0s := strings.Split(e0.Error(), "transaction:")
	e1s := strings.Split(e1.Error(), "transaction:")
	e2s := strings.Split(e2.Error(), "transaction:")

	assert.NotEqual(t, e0s[1], e1s[1], "transaction id shoul be different")
	assert.Equal(t, e1s[1], e2s[1], "transaction id should be equal")
}
