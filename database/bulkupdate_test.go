package database

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	assert1 "github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestBulkUpdate(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	configuration := &InstanceConfig{
		Log: log.Global(),
	}
	statement := transaction.NewMockNamedStatement(ctrl)
	tx := transaction.NewMockTransaction(ctrl)
	client := NewMockClient(ctrl)
	db := &transactable{db: client}
	bulkUpdate := &bulkUpdate[*TestRecord]{
		bulkOperation: &bulkOperation[*TestRecord]{
			tx:            tx,
			db:            db,
			configuration: configuration,
		},
		columns: []qb.TableField{MetaTestRecord.Name},
	}
	expected := &TestRecord{
		ID:   generator.ID("test"),
		Name: generator.String(32),
	}

	tx.EXPECT().PrepareNamed(
		"UPDATE `test_record` SET  `test_record`.`name` = :name "+
			"WHERE `test_record`.`id` = :id").Return(statement, nil)
	statement.EXPECT().Exec(expected)
	statement.EXPECT().Close().Return(nil)
	tx.EXPECT().Commit().Return(nil)
	bulkUpdate.Update(expected)
	_, err := bulkUpdate.Commit()
	assert.NoError(err)
}
