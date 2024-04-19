package database

import (
	"fmt"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	assert1 "github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

type TestRecord struct {
	ID   string
	Name string
}

func (tc *TestRecord) Initialize() {
	tc.ID = generator.ID("test")
}

func (tc *TestRecord) PrimaryKey() record.PrimaryKeyValue {
	return record.NewPrimaryKey(tc.ID)
}

func (tc *TestRecord) Key() string {
	return tc.ID
}

func (tc *TestRecord) Meta() qb.Table {
	return MetaTestRecord
}

type metaTestRecord struct {
	alias      string
	ID         qb.TableField
	Name       qb.TableField
	allColumns qb.TableField
}

func (t *metaTestRecord) GetName() string {
	return "test_record"
}

func (t *metaTestRecord) GetAlias() string {
	return t.alias
}

func (t *metaTestRecord) PrimaryKey() qb.TableField {
	return t.ID
}

func (t *metaTestRecord) SortBy() (qb.TableField, qb.OrderDirection) {
	return t.ID, qb.Ascending
}

func (t *metaTestRecord) AllColumns() qb.TableField {
	return t.allColumns
}

func (t *metaTestRecord) ReadColumns() []qb.TableField {
	return []qb.TableField{
		t.ID,
		t.Name,
	}
}

func (t *metaTestRecord) WriteColumns() []qb.TableField {
	return t.ReadColumns()
}

func (t *metaTestRecord) Alias(alias string) *metaTestRecord {
	return &metaTestRecord{
		alias:      alias,
		ID:         qb.TableField{Name: "id", Table: alias},
		Name:       qb.TableField{Name: "name", Table: alias},
		allColumns: qb.TableField{Name: "*", Table: alias},
	}
}

var MetaTestRecord = (&metaTestRecord{}).Alias("test_record")

type TestRecordCollection []*TestRecord

type countMatcher struct {
	count int32
}

func (matcher *countMatcher) Matches(x interface{}) bool {
	rows, ok := x.(*[]*qb.RowCount)
	if !ok {
		return false
	}
	*rows = append(*rows, &qb.RowCount{Count: int(matcher.count)})
	return true
}

func (matcher *countMatcher) String() string {
	return fmt.Sprintf("countMatcher(%d)", matcher.count)
}

type queryMatcher struct {
	t   *testing.T
	sql string
}

func (matcher *queryMatcher) Matches(x interface{}) bool {
	query, ok := x.(*qb.SelectQuery)
	if !ok {
		// only return false if we got an unexpected type, otherwise let
		// the assert take care of failure and messaging
		return false
	}
	sql, _, err := query.SQL(qb.NoLimit, 0)
	assert1.NoError(matcher.t, err)
	assert1.Equal(matcher.t, matcher.sql, sql)
	return true
}

func (matcher *queryMatcher) String() string {
	return "*qb.SelectQuery"
}

func Test_SelectWithTotal(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	api := NewMockAPI(ctrl)

	var (
		err    error
		total  int
		target TestRecordCollection
		limit = 0
		offset = 0
	)

	query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)

	//limit equal zero 
	api.EXPECT().Count(MetaTestRecord, query).Return(int32(2), nil)
	
	_, total, err = SelectWithTotal(api, MetaTestRecord, target, query, limit, offset)

	assert.NoError(err)
	assert.Equal(2, total)

	//total equal zero 
	limit = 1 //change limit

	api.EXPECT().Count(MetaTestRecord, query).Return(int32(0), nil)
	
	_, total, err = SelectWithTotal(api, MetaTestRecord, target, query, limit, offset)

	assert.NoError(err)
	assert.Equal(0, total)

	//total equal 1 
	api.EXPECT().Count(MetaTestRecord, query).Return(int32(1), nil)
	api.EXPECT().Select(&target, query, record.NewListOptions(limit, offset)).Return(nil)

	_, total, err = SelectWithTotal(api, MetaTestRecord, target, query, limit, offset)
	
	assert.NoError(err)
	assert.Equal(1, total)

	//error cases
	expected := generator.String(20)

	api.EXPECT().Count(MetaTestRecord, query).Return(int32(0), errors.New(expected)) 

	_, total, err = SelectWithTotal(api, MetaTestRecord, target, query, limit, offset)
	assert.Equal(0, total)
	assert.EqualError(err, expected)

	api.EXPECT().Count(MetaTestRecord, query).Return(int32(1), nil)
	api.EXPECT().Select(&target, query, record.NewListOptions(limit, offset)).Return(errors.New(expected))

	_, total, err = SelectWithTotal(api, MetaTestRecord, target, query, limit, offset)
	assert.EqualError(err, expected)
	assert.Equal(0, total)
}

func Test_Begin(t *testing.T){
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	transaction := transaction.NewMockTransaction(ctrl)
	var err error 
	api := &api{
		tx:            transaction,
		configuration: &InstanceConfig{MaxLimit: 100},
	}
	err = api.Begin()
	assert.NoError(err)
}


func Test_Rollback(t *testing.T){
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	transaction := transaction.NewMockTransaction(ctrl)
	var err error 
	api := &api{
		tx:            transaction,
		configuration: &InstanceConfig{MaxLimit: 100},
	}

	transaction.EXPECT().Rollback().Return(nil)

	err = api.Rollback()

	assert.NoError(err)
}

func Test_Rollback_transactionIsNil(t *testing.T){
	assert := assert1.New(t)

	api := &api{
		tx:            nil,
		configuration: &InstanceConfig{MaxLimit: 100},
	}

	err := api.Rollback()
	assert.EqualError(err, ErrMissingTransaction.Error())
}

func Test_Commit(t *testing.T){
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	transaction := transaction.NewMockTransaction(ctrl)

	var err error 

	api := &api{
		tx:            transaction,
		configuration: &InstanceConfig{MaxLimit: 100},
	}

	transaction.EXPECT().Commit().Return(nil)

	err = api.Commit()

	assert.NoError(err)
}

func Test_Commit_transactionIsNil(t *testing.T){
	assert := assert1.New(t)

	api := &api{
		tx:            nil,
		configuration: &InstanceConfig{MaxLimit: 100},
	}

	err := api.Commit()
	assert.EqualError(err, ErrMissingTransaction.Error())
}

func Test_ApiMethods(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	transaction := transaction.NewMockTransaction(ctrl)

	api := &api{
		tx:            transaction,
		configuration: &InstanceConfig{MaxLimit: 100},
	}

	target := &TestRecord{ID: "2", Name: "test"}
	listOptions := &record.ListOptions{Limit: 1, Offset: 0}
	query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)
	conditionalExpression := qb.FieldComparison(MetaTestRecord.Name, qb.Equal, "")

	transaction.EXPECT().Create(target)
	api.Create(target)

	transaction.EXPECT().Read(target, target.PrimaryKey()).Return(nil)
	api.Read(target, target.PrimaryKey())
	
	transaction.EXPECT().ReadOneWhere(target, conditionalExpression).Return(nil)
	api.ReadOneWhere(target, conditionalExpression)

	transaction.EXPECT().Select(target, query, *listOptions).Return(nil)
	api.Select(target, query, listOptions)

	transaction.EXPECT().ListWhere(target, target, conditionalExpression, *listOptions).Return(nil)
	api.ListWhere(target, target, conditionalExpression, listOptions)

	transaction.EXPECT().Update(target).Return(nil)
	api.Update(target)

	transaction.EXPECT().UpdateWhere(target, conditionalExpression, qb.FieldValue{Field: MetaTestRecord.ID, Value: 2 }).Return(int64(0), nil)
	total, err := api.UpdateWhere(target, conditionalExpression, qb.FieldValue{Field: MetaTestRecord.ID, Value: 2 })

	assert.Equal(int64(0), total)
	assert.NoError(err)

	transaction.EXPECT().Delete(target).Return(nil)
	api.Delete(target)

	transaction.EXPECT().DeleteWhere(target, conditionalExpression).Return(nil)
	api.DeleteWhere(target, conditionalExpression)
}

func Test_database_enforceLimits(t *testing.T) {
	var tests = []struct {
		name          string
		maxQueryLimit uint
		options       *record.ListOptions
		expected      *record.ListOptions
	}{
		{
			name:          "no limit",
			maxQueryLimit: 0,
			options:       &record.ListOptions{Limit: 100, Offset: 0},
			expected:      &record.ListOptions{Limit: 100, Offset: 0},
		},
		{
			name:          "limit enforced",
			maxQueryLimit: 10,
			options:       &record.ListOptions{Limit: 100, Offset: 0},
			expected:      &record.ListOptions{Limit: 10, Offset: 0},
		},
		{
			name:          "nil gets defaults",
			maxQueryLimit: 20,
			options:       nil,
			expected:      &record.ListOptions{Limit: 20, Offset: 0},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert1.New(t)
			conf := &InstanceConfig{}
			conf.MaxLimit = tc.maxQueryLimit
			database := &api{configuration: conf}
			actual := database.enforceLimits(tc.options)
			assert.Equal(tc.expected, actual)
		})
	}
}

func Test_api_Count(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	transaction := transaction.NewMockTransaction(ctrl)
	api := &api{
		tx:            transaction,
		configuration: &InstanceConfig{MaxLimit: 100},
	}
	query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)
	expected := generator.Int32()
	transaction.EXPECT().Select(&countMatcher{count: expected},
		&queryMatcher{t: t, sql: "SELECT COUNT(*) as count FROM " +
			"`test_record` AS `test_record`"},
		record.ListOptions{Limit: 1, Offset: 0},
	).Return(nil)
	actual, err := api.Count(MetaTestRecord, query)
	assert.NoError(err)
	assert.Equal(expected, actual)
}

func Test_api_Count_zero(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	transaction := transaction.NewMockTransaction(ctrl)
	api := &api{
		tx:            transaction,
		configuration: &InstanceConfig{MaxLimit: 100},
	}
	query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)
	transaction.EXPECT().Select(gomock.Any(),
		&queryMatcher{t: t, sql: "SELECT COUNT(*) as count FROM " +
			"`test_record` AS `test_record`"},
		record.ListOptions{Limit: 1, Offset: 0},
	).Return(nil)
	actual, err := api.Count(MetaTestRecord, query)
	assert.NoError(err)
	assert.Equal(int32(0), actual)
}

func Test_api_Count_error(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	transaction := transaction.NewMockTransaction(ctrl)
	api := &api{
		tx:            transaction,
		configuration: &InstanceConfig{MaxLimit: 100},
	}
	expected := generator.String(20)
	query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)
	transaction.EXPECT().Select(gomock.Any(),
		&queryMatcher{t: t, sql: "SELECT COUNT(*) as count FROM " +
			"`test_record` AS `test_record`"},
		record.ListOptions{Limit: 1, Offset: 0},
	).Return(errors.New(expected))
	_, err := api.Count(MetaTestRecord, query)
	assert.EqualError(err, expected)
}

func Test_api_CountWhere_nil(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	transaction := transaction.NewMockTransaction(ctrl)
	api := &api{
		tx:            transaction,
		configuration: &InstanceConfig{MaxLimit: 100},
	}

	expected := generator.Int32()
	transaction.EXPECT().Select(&countMatcher{count: expected},
		&queryMatcher{t: t, sql: "SELECT COUNT(*) as count FROM " +
			"`test_record` AS `test_record`"},
		record.ListOptions{Limit: 1, Offset: 0},
	).Return(nil)
	actual, err := api.CountWhere(MetaTestRecord, nil)
	assert.NoError(err)
	assert.Equal(expected, actual)
}

func Test_api_CountWhere(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	transaction := transaction.NewMockTransaction(ctrl)
	api := &api{
		tx:            transaction,
		configuration: &InstanceConfig{MaxLimit: 100},
	}

	expected := generator.Int32()
	transaction.EXPECT().Select(&countMatcher{count: expected},
		&queryMatcher{t: t, sql: "SELECT COUNT(*) as count FROM `test_record` AS" +
			" `test_record` WHERE `test_record`.`name` = ?"},
		record.ListOptions{Limit: 1, Offset: 0},
	).Return(nil)
	actual, err := api.CountWhere(MetaTestRecord,
		qb.FieldComparison(MetaTestRecord.Name, qb.Equal, ""))
	assert.NoError(err)
	assert.Equal(expected, actual)
}

// TODO: [COR-587] finish tests for API
