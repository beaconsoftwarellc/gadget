package database

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/golang/mock/gomock"
	assert1 "github.com/stretchr/testify/assert"
)

func getMocks(t *testing.T) *MockAPI {
	ctrl := gomock.NewController(t)
	dbMock := NewMockAPI(ctrl)
	return dbMock
}

func Test_Count_No_Rows(t *testing.T) {
	assert := assert1.New(t)
	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).Return(nil)

	_, actual := Count(mockDB, Action, qb.Select(Action.ID).From(Action))
	assert.EqualError(actual, "[COM.DB.1] row count query execution failed (no rows)")
}

func Test_Count_DB_Error(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))

	mockDB := getMocks(t)
	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).Return(expected)

	_, actual := Count(mockDB, Action, qb.Select(Action.ID).From(Action))
	assert.EqualError(actual, expected.Error())
}

func Test_Count(t *testing.T) {
	assert := assert1.New(t)
	expected := 13
	clauseValue := generator.String(20)

	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).
		Do(func(target, query interface{}) interface{} {

			cTarget, _ := target.(*([]*qb.RowCount))
			*cTarget = append(*cTarget, &qb.RowCount{Count: expected})

			cQuery := query.(*qb.SelectQuery)
			query, arguments, err := cQuery.SQL(0, 0)

			assert.NoError(err)
			assert.Equal([]interface{}{clauseValue}, arguments)
			assert.Equal(query, "SELECT COUNT(*) as count FROM `action` AS `action` INNER JOIN `action` AS `action` ON `action`.`id` = `action`.`id` WHERE `action`.`name` = ?")

			return nil
		}).Return(nil)

	q := qb.Select(Action.ID).From(Action)
	q.InnerJoin(Action).On(Action.ID, qb.Equal, Action.ID)
	q.Where(Action.Name.Equal(clauseValue))

	actual, err := Count(mockDB, Action, q)
	assert.NoError(err)
	assert.Equal(expected, int(actual))
}

func Test_TableCount_No_Rows(t *testing.T) {
	assert := assert1.New(t)
	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).Return(nil)

	_, actual := CountTableRows(mockDB, Action)
	assert.EqualError(actual, "[COM.DB.1] row count query execution failed (no rows)")
}

func Test_TableCount_DB_Error(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))

	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).Return(expected)

	_, actual := CountTableRows(mockDB, Action)
	assert.EqualError(actual, expected.Error())
}

func Test_TableCount(t *testing.T) {
	assert := assert1.New(t)
	expected := 13

	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).
		Do(func(target, query interface{}) interface{} {

			cTarget, _ := target.(*([]*qb.RowCount))
			*cTarget = append(*cTarget, &qb.RowCount{Count: expected})

			cQuery := query.(*qb.SelectQuery)
			query, arguments, err := cQuery.SQL(0, 0)

			assert.NoError(err)
			assert.Equal([]interface{}([]interface{}(nil)), arguments)
			assert.Equal(query, "SELECT COUNT(*) as count FROM `action` AS `action`")

			return nil
		}).Return(nil)

	actual, err := CountTableRows(mockDB, Action)
	assert.NoError(err)
	assert.Equal(expected, int(actual))
}

func Test_CountWhere_No_Rows(t *testing.T) {
	assert := assert1.New(t)
	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).Return(nil)

	_, actual := CountWhere(mockDB, Action, Action.Name.Equal("name"))
	assert.EqualError(actual, "[COM.DB.1] row count query execution failed (no rows)")
}

func Test_CountWhere_DB_Error(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))

	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).Return(expected)

	_, actual := CountWhere(mockDB, Action, Action.Name.Equal("name"))
	assert.EqualError(actual, expected.Error())
}

func Test_CountWhere(t *testing.T) {
	assert := assert1.New(t)
	expected := 13
	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).
		Do(func(target, query interface{}) interface{} {

			cTarget, _ := target.(*([]*qb.RowCount))
			*cTarget = append(*cTarget, &qb.RowCount{Count: expected})

			cQuery := query.(*qb.SelectQuery)
			query, arguments, err := cQuery.SQL(0, 0)

			assert.NoError(err)
			assert.Equal([]interface{}{"name"}, arguments)
			assert.Equal(query, "SELECT COUNT(*) as count FROM `action` AS `action` WHERE `action`.`name` = ?")

			return nil
		}).Return(nil)

	actual, err := CountWhere(mockDB, Action, Action.Name.Equal("name"))
	assert.NoError(err)
	assert.Equal(expected, int(actual))
}
