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

type testRecord struct {
	alias      string
	ID         qb.TableField
	Name       qb.TableField
	allColumns qb.TableField
}

func (t *testRecord) GetName() string {
	return "test_record"
}

func (t *testRecord) GetAlias() string {
	return t.alias
}

func (t *testRecord) PrimaryKey() qb.TableField {
	return t.ID
}

func (t *testRecord) SortBy() (qb.TableField, qb.OrderDirection) {
	return t.ID, qb.Ascending
}

func (t *testRecord) AllColumns() qb.TableField {
	return t.allColumns
}

func (t *testRecord) ReadColumns() []qb.TableField {
	return []qb.TableField{
		t.ID,
		t.Name,
	}
}

func (t *testRecord) WriteColumns() []qb.TableField {
	return t.ReadColumns()
}

func (t *testRecord) Alias(alias string) *testRecord {
	return &testRecord{
		alias:      alias,
		ID:         qb.TableField{Name: "id", Table: alias},
		Name:       qb.TableField{Name: "name", Table: alias},
		allColumns: qb.TableField{Name: "*", Table: alias},
	}
}

var TestRecord = (&testRecord{}).Alias("test_record")

func Test_Count_No_Rows(t *testing.T) {
	assert := assert1.New(t)
	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).Return(nil)

	_, actual := Count(mockDB, TestRecord, qb.Select(TestRecord.ID).From(TestRecord))
	assert.EqualError(actual, "[COM.DB.1] row count query execution failed (no rows)")
}

func Test_Count_DB_Error(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))

	mockDB := getMocks(t)
	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).Return(expected)

	_, actual := Count(mockDB, TestRecord, qb.Select(TestRecord.ID).From(TestRecord))
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
			assert.Equal(query, "SELECT COUNT(*) as count FROM `test_record` AS `test_record` INNER JOIN `test_record` AS `test_record` ON `test_record`.`id` = `test_record`.`id` WHERE `test_record`.`name` = ?")

			return nil
		}).Return(nil)

	q := qb.Select(TestRecord.ID).From(TestRecord)
	q.InnerJoin(TestRecord).On(TestRecord.ID, qb.Equal, TestRecord.ID)
	q.Where(TestRecord.Name.Equal(clauseValue))

	actual, err := Count(mockDB, TestRecord, q)
	assert.NoError(err)
	assert.Equal(expected, int(actual))
}

func Test_TableCount_No_Rows(t *testing.T) {
	assert := assert1.New(t)
	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).Return(nil)

	_, actual := CountTableRows(mockDB, TestRecord)
	assert.EqualError(actual, "[COM.DB.1] row count query execution failed (no rows)")
}

func Test_TableCount_DB_Error(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))

	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).Return(expected)

	_, actual := CountTableRows(mockDB, TestRecord)
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
			assert.Equal(query, "SELECT COUNT(*) as count FROM `test_record` AS `test_record`")

			return nil
		}).Return(nil)

	actual, err := CountTableRows(mockDB, TestRecord)
	assert.NoError(err)
	assert.Equal(expected, int(actual))
}

func Test_CountWhere_No_Rows(t *testing.T) {
	assert := assert1.New(t)
	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).Return(nil)

	_, actual := CountWhere(mockDB, TestRecord, TestRecord.Name.Equal("name"))
	assert.EqualError(actual, "[COM.DB.1] row count query execution failed (no rows)")
}

func Test_CountWhere_DB_Error(t *testing.T) {
	assert := assert1.New(t)
	expected := errors.New(generator.String(20))

	mockDB := getMocks(t)

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any()).Return(expected)

	_, actual := CountWhere(mockDB, TestRecord, TestRecord.Name.Equal("name"))
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
			assert.Equal(query, "SELECT COUNT(*) as count FROM `test_record` AS `test_record` WHERE `test_record`.`name` = ?")

			return nil
		}).Return(nil)

	actual, err := CountWhere(mockDB, TestRecord, TestRecord.Name.Equal("name"))
	assert.NoError(err)
	assert.Equal(expected, int(actual))
}
