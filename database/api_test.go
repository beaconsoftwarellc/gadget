package database

import (
	"fmt"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	_require "github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
	_require.NoError(matcher.t, err)
	_require.Equal(matcher.t, matcher.sql, sql)
	return true
}

func (matcher *queryMatcher) String() string {
	return "*qb.SelectQuery"
}

func Test_SelectWithTotal(t *testing.T) {

	var (
		ctrl   = gomock.NewController(t)
		api    = NewMockAPI(ctrl)
		target TestRecordCollection
		limit  = 0
		offset = 0
	)

	query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)

	t.Run("limit equal zero", func(t *testing.T) {
		require := _require.New(t)
		api.EXPECT().Count(MetaTestRecord, query).Return(int32(2), nil)
		_, total, err := SelectWithTotal(api, MetaTestRecord, target, query, limit, offset)
		require.NoError(err)
		require.Equal(2, total)
	})

	t.Run("total equal zero", func(t *testing.T) {
		require := _require.New(t)
		limit = 1 // change limit
		api.EXPECT().Count(MetaTestRecord, query).Return(int32(0), nil)
		_, total, err := SelectWithTotal(api, MetaTestRecord, target, query, limit, offset)
		require.NoError(err)
		require.Equal(0, total)
	})

	t.Run("total equal 1", func(t *testing.T) {
		require := _require.New(t)
		api.EXPECT().Count(MetaTestRecord, query).Return(int32(1), nil)
		api.EXPECT().Select(&target, query, record.NewListOptions(limit, offset)).Return(nil)
		_, total, err := SelectWithTotal(api, MetaTestRecord, target, query, limit, offset)
		require.NoError(err)
		require.Equal(1, total)
	})

	t.Run("error case: Count error", func(t *testing.T) {
		require := _require.New(t)
		expected := generator.String(20)
		api.EXPECT().Count(MetaTestRecord, query).Return(int32(0), errors.New(expected))
		_, total, err := SelectWithTotal(api, MetaTestRecord, target, query, limit, offset)
		require.Equal(0, total)
		require.EqualError(err, expected)
	})

	t.Run("error case: Select error", func(t *testing.T) {
		require := _require.New(t)
		expected := generator.String(20)
		api.EXPECT().Count(MetaTestRecord, query).Return(int32(1), nil)
		api.EXPECT().Select(&target, query, record.NewListOptions(limit, offset)).Return(errors.New(expected))
		_, total, err := SelectWithTotal(api, MetaTestRecord, target, query, limit, offset)
		require.EqualError(err, expected)
		require.Equal(0, total)
	})
}

func Test_Begin(t *testing.T) {
	require := _require.New(t)
	ctrl := gomock.NewController(t)
	transaction := transaction.NewMockTransaction(ctrl)
	var err error
	api := &api{
		tx:            transaction,
		configuration: &InstanceConfig{MaxLimit: 100},
	}
	err = api.Begin()
	require.NoError(err)
}

func Test_Rollback(t *testing.T) {
	require := _require.New(t)
	ctrl := gomock.NewController(t)
	transaction := transaction.NewMockTransaction(ctrl)
	var err error
	api := &api{
		tx:            transaction,
		configuration: &InstanceConfig{MaxLimit: 100},
	}

	transaction.EXPECT().Rollback().Return(nil)

	err = api.Rollback()

	require.NoError(err)
}

func Test_Rollback_transactionIsNil(t *testing.T) {
	require := _require.New(t)

	api := &api{
		tx:            nil,
		configuration: &InstanceConfig{MaxLimit: 100},
	}

	err := api.Rollback()
	require.EqualError(err, ErrMissingTransaction.Error())
}

func Test_Commit(t *testing.T) {
	require := _require.New(t)
	ctrl := gomock.NewController(t)
	transaction := transaction.NewMockTransaction(ctrl)

	var err error

	api := &api{
		tx:            transaction,
		configuration: &InstanceConfig{MaxLimit: 100},
	}

	transaction.EXPECT().Commit().Return(nil)

	err = api.Commit()

	require.NoError(err)
}

func Test_Commit_transactionIsNil(t *testing.T) {
	require := _require.New(t)

	api := &api{
		tx:            nil,
		configuration: &InstanceConfig{MaxLimit: 100},
	}

	err := api.Commit()
	require.EqualError(err, ErrMissingTransaction.Error())
}

func Test_ApiMethods(t *testing.T) {
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

	t.Run("Create", func(t *testing.T) {
		require := _require.New(t)
		transaction.EXPECT().Create(target).Return(nil)
		err := api.Create(target)
		require.NoError(err)
	})

	t.Run("Read", func(t *testing.T) {
		require := _require.New(t)
		transaction.EXPECT().Read(target, target.PrimaryKey()).Return(nil)
		err := api.Read(target, target.PrimaryKey())
		require.NoError(err)
	})

	t.Run("ReadOneWhere", func(t *testing.T) {
		require := _require.New(t)
		transaction.EXPECT().ReadOneWhere(target, conditionalExpression).Return(nil)
		err := api.ReadOneWhere(target, conditionalExpression)
		require.NoError(err)
	})

	t.Run("Select", func(t *testing.T) {
		require := _require.New(t)
		transaction.EXPECT().Select(target, query, *listOptions).Return(nil)
		err := api.Select(target, query, listOptions)
		require.NoError(err)
	})

	t.Run("ListWhere", func(t *testing.T) {
		require := _require.New(t)
		transaction.EXPECT().ListWhere(target, target, conditionalExpression, *listOptions).Return(nil)
		err := api.ListWhere(target, target, conditionalExpression, listOptions)
		require.NoError(err)
	})

	t.Run("Update", func(t *testing.T) {
		require := _require.New(t)
		transaction.EXPECT().Update(target).Return(nil)
		err := api.Update(target)
		require.NoError(err)
	})

	t.Run("Delete", func(t *testing.T) {
		require := _require.New(t)
		transaction.EXPECT().Delete(target).Return(nil)
		err := api.Delete(target)
		require.NoError(err)
	})

	t.Run("DeleteWhere", func(t *testing.T) {
		require := _require.New(t)
		transaction.EXPECT().DeleteWhere(target, conditionalExpression).Return(nil)
		err := api.DeleteWhere(target, conditionalExpression)
		require.NoError(err)
	})

	t.Run("UpdateWhere", func(t *testing.T) {
		require := _require.New(t)
		transaction.EXPECT().UpdateWhere(target, conditionalExpression, qb.FieldValue{Field: MetaTestRecord.ID, Value: 2}).Return(int64(0), nil)
		total, err := api.UpdateWhere(target, conditionalExpression, qb.FieldValue{Field: MetaTestRecord.ID, Value: 2})
		require.NoError(err)
		require.Equal(int64(0), total)
	})
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
			require := _require.New(t)
			conf := &InstanceConfig{}
			conf.MaxLimit = tc.maxQueryLimit
			database := &api{configuration: conf}
			actual := database.enforceLimits(tc.options)
			require.Equal(tc.expected, actual)
		})
	}
}

func Test_api_Count(t *testing.T) {
	require := _require.New(t)
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
	require.NoError(err)
	require.Equal(expected, actual)
}

func Test_api_Count_zero(t *testing.T) {
	require := _require.New(t)
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
	require.NoError(err)
	require.Equal(int32(0), actual)
}

func Test_api_Count_error(t *testing.T) {
	require := _require.New(t)
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
	require.EqualError(err, expected)
}

func Test_api_CountWhere_nil(t *testing.T) {
	require := _require.New(t)
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
	require.NoError(err)
	require.Equal(expected, actual)
}

func Test_api_CountWhere(t *testing.T) {
	require := _require.New(t)
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
	require.NoError(err)
	require.Equal(expected, actual)
}
