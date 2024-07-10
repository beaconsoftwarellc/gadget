package database

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/jmoiron/sqlx"
	assert1 "github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBulkCreateReset(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)

	configuration := &InstanceConfig{
		Log: log.Global(),
	}
	transaction := transaction.NewMockTransaction(ctrl)
	client := NewMockClient(ctrl)
	db := &transactable{db: client}

	bulkCreate := &bulkCreate[*TestRecord]{
		bulkOperation: &bulkOperation[*TestRecord]{
			tx:            transaction,
			db:            db,
			configuration: configuration,
		},
	}

	// transaction is not nil should fail
	actual := bulkCreate.Reset()
	assert.EqualError(actual, "transaction should be committed or"+
		" rolled back prior to calling Reset")

	bulkCreate.pending = []*TestRecord{
		{
			ID:   generator.ID("test"),
			Name: generator.String(10),
		},
	}
	bulkCreate.tx = nil

	tx := &sqlx.Tx{}
	client.EXPECT().Beginx().Return(tx, nil)
	actual = bulkCreate.Reset()
	assert.NoError(actual)
	assert.Empty(bulkCreate.pending)
	assert.NotNil(bulkCreate.tx)
}

func TestBulkCreateCreate(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)

	configuration := &InstanceConfig{
		Log: log.Global(),
	}
	transaction := transaction.NewMockTransaction(ctrl)
	client := NewMockClient(ctrl)
	db := &transactable{db: client}

	bulkCreate := &bulkCreate[*TestRecord]{
		bulkOperation: &bulkOperation[*TestRecord]{
			tx:            transaction,
			db:            db,
			configuration: configuration,
		},
	}

	testRecord := &TestRecord{Name: generator.String(32)}
	bulkCreate.Create(testRecord)
	assert.Equal(1, len(bulkCreate.pending))
	assert.NotEmpty(testRecord.ID)
	assert.Equal(testRecord.ID, bulkCreate.pending[0].ID)

	testRecord1 := &TestRecord{Name: generator.String(32)}
	testRecord2 := &TestRecord{Name: generator.String(32)}
	bulkCreate.Create(testRecord1, testRecord2)
	assert.Equal(3, len(bulkCreate.pending))
	assert.NotEmpty(testRecord1.ID)
	assert.NotEmpty(testRecord2.ID)
	assert.Equal(testRecord.ID, bulkCreate.pending[0].ID)
	assert.Equal(testRecord1.ID, bulkCreate.pending[1].ID)
	assert.Equal(testRecord2.ID, bulkCreate.pending[2].ID)
}

func TestBulkCreateCommit(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)

	configuration := &InstanceConfig{
		Log: log.Global(),
	}
	implementation := transaction.NewMockImplementation(ctrl)
	transaction := transaction.NewMockTransaction(ctrl)
	transaction.EXPECT().Implementation().Return(implementation).AnyTimes()
	client := NewMockClient(ctrl)
	db := &transactable{db: client}

	bulkCreate := &bulkCreate[*TestRecord]{
		bulkOperation: &bulkOperation[*TestRecord]{
			tx:            transaction,
			db:            db,
			configuration: configuration,
		},
	}
	testRecord := &TestRecord{Name: generator.String(32)}
	testRecord1 := &TestRecord{Name: generator.String(32)}
	bulkCreate.Create(testRecord, testRecord1)
	implementation.EXPECT().NamedExec("INSERT INTO `test_record` "+
		"(`test_record`.`id`, `test_record`.`name`) VALUES (:id, :name)",
		bulkCreate.pending).Return(nil, nil)
	transaction.EXPECT().Commit().Return(nil)
	_, actualErr := bulkCreate.Commit()
	assert.NoError(actualErr)
}

func TestBulkCreateCommitUpsert(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)

	configuration := &InstanceConfig{
		Log: log.Global(),
	}
	implementation := transaction.NewMockImplementation(ctrl)
	transaction := transaction.NewMockTransaction(ctrl)
	transaction.EXPECT().Implementation().Return(implementation).AnyTimes()
	client := NewMockClient(ctrl)
	db := &transactable{db: client}

	bulkCreate := &bulkCreate[*TestRecord]{
		bulkOperation: &bulkOperation[*TestRecord]{
			tx:            transaction,
			db:            db,
			configuration: configuration,
		},
		upsert: true,
	}
	testRecord := &TestRecord{Name: generator.String(32)}
	testRecord1 := &TestRecord{Name: generator.String(32)}
	bulkCreate.Create(testRecord, testRecord1)
	implementation.EXPECT().NamedExec("INSERT INTO `test_record` "+
		"(`test_record`.`id`, `test_record`.`name`) VALUES (:id, :name) "+
		"ON DUPLICATE KEY UPDATE `test_record`.`id` = VALUES(`test_record`.`id`), "+
		"`test_record`.`name` = VALUES(`test_record`.`name`)",
		bulkCreate.pending).Return(nil, nil)
	transaction.EXPECT().Commit().Return(nil)
	_, actualErr := bulkCreate.Commit()
	assert.NoError(actualErr)
}

func TestBulkCreateRollback(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)

	configuration := &InstanceConfig{
		Log: log.Global(),
	}
	transaction := transaction.NewMockTransaction(ctrl)
	client := NewMockClient(ctrl)
	db := &transactable{db: client}

	bulkCreate := &bulkCreate[*TestRecord]{
		bulkOperation: &bulkOperation[*TestRecord]{
			tx:            transaction,
			db:            db,
			configuration: configuration,
		},
	}
	testRecord := &TestRecord{Name: generator.String(32)}
	bulkCreate.Create(testRecord)

	expected := generator.ID("err")
	transaction.EXPECT().Rollback().Return(errors.New(expected))
	actual := bulkCreate.Rollback()
	assert.EqualError(actual, expected)
	assert.Nil(bulkCreate.tx)
	assert.Empty(bulkCreate.pending)
}
