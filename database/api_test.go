package database

import (
	"fmt"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
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

func Test_database_enforceLimits(t *testing.T) {
	var tests = []struct {
		name          string
		maxQueryLimit uint
		options       record.LimitOffset
		expected      record.LimitOffset
	}{
		{
			name:          "no limit",
			maxQueryLimit: 0,
			options:       record.NewLimitOffset[int]().SetLimit(100).SetOffset(0),
			expected:      record.NewLimitOffset[int]().SetLimit(0).SetOffset(0),
		},
		{
			name:          "limit enforced",
			maxQueryLimit: 10,
			options:       record.NewLimitOffset[int]().SetLimit(100).SetOffset(0),
			expected:      record.NewLimitOffset[int]().SetLimit(10).SetOffset(0),
		},
		{
			name:          "nil gets defaults",
			maxQueryLimit: 20,
			options:       nil,
			expected:      record.NewLimitOffset[int]().SetLimit(20).SetOffset(0),
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
		record.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
	).Return(nil)
	actual, err := api.Count(MetaTestRecord, query)
	assert.NoError(err)
	assert.Equal(expected, actual)
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
		record.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
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
		record.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
	).Return(nil)
	actual, err := api.CountWhere(MetaTestRecord,
		qb.FieldComparison(MetaTestRecord.Name, qb.Equal, ""))
	assert.NoError(err)
	assert.Equal(expected, actual)
}

// TODO: [COR-587] finish tests for API
