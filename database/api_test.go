package database

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAPISlowLog(t *testing.T) {
	spec := newSpecification()
	logger := log.NewMockLogger(gomock.NewController(t))

	record := &TestRecord{Name: generator.Name()}

	api := NewAPI(spec.DB, nil, logger)

	logger.EXPECT().Debug(gomock.Any()).Return(nil)
	err := api.Create(record)
	require.NoError(t, err)

	api.SetSlowQueryDuration(0)
	logger.EXPECT().Error(gomock.Any()).Return(nil)
	record = &TestRecord{Name: generator.Name()}
	err = api.Create(record)
	require.NoError(t, err)
}
