package database

import (
	"fmt"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/jmoiron/sqlx"
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

func (matcher *queryMatcher) Matches(x any) bool {
	query, ok := x.(*qb.SelectQuery)
	if !ok {
		// only return false if we got an unexpected type, otherwise let
		// the assert take care of failure and messaging
		return false
	}
	sql, _, err := query.SQL(nil)
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
		options       qb.LimitOffset
		expected      qb.LimitOffset
	}{
		{
			name:          "no limit",
			maxQueryLimit: 0,
			options:       qb.NewLimitOffset[int]().SetLimit(100).SetOffset(0),
			expected:      qb.NewLimitOffset[int]().SetLimit(100).SetOffset(0),
		},
		{
			name:          "limit enforced",
			maxQueryLimit: 10,
			options:       qb.NewLimitOffset[int]().SetLimit(100).SetOffset(0),
			expected:      qb.NewLimitOffset[int]().SetLimit(10).SetOffset(0),
		},
		{
			name:          "nil gets defaults",
			maxQueryLimit: 20,
			options:       nil,
			expected:      qb.NewLimitOffset[int]().SetLimit(20).SetOffset(0),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert1.New(t)
			conf := &InstanceConfig{}
			conf.MaxLimit = tc.maxQueryLimit
			database := &api{configuration: conf}
			actual := database.enforceLimits(tc.options)
			assert.Equal(tc.expected.Limit(), actual.Limit())
			assert.Equal(tc.expected.Offset(), actual.Offset())
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
		qb.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
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
		qb.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
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
		qb.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
	).Return(nil)
	actual, err := api.CountWhere(MetaTestRecord,
		qb.FieldComparison(MetaTestRecord.Name, qb.Equal, ""))
	assert.NoError(err)
	assert.Equal(expected, actual)
}

type sumMatcher struct {
	sum *int
}

func (matcher *sumMatcher) Matches(x interface{}) bool {
	rows, ok := x.(*[]*qb.SumResult)
	if !ok {
		return false
	}
	*rows = append(*rows, &qb.SumResult{Sum: matcher.sum})
	return true
}

func (matcher *sumMatcher) String() string {
	if matcher.sum == nil {
		return "sumMatcher(nil)"
	}
	return fmt.Sprintf("sumMatcher(%d)", *matcher.sum)
}

type emptyCountMatcher struct{}

func (matcher *emptyCountMatcher) Matches(x interface{}) bool {
	_, ok := x.(*[]*qb.RowCount)
	return ok
}

func (matcher *emptyCountMatcher) String() string {
	return "emptyCountMatcher"
}

type emptySumMatcher struct{}

func (matcher *emptySumMatcher) Matches(x interface{}) bool {
	_, ok := x.(*[]*qb.SumResult)
	return ok
}

func (matcher *emptySumMatcher) String() string {
	return "emptySumMatcher"
}

func Test_SelectWithTotal(t *testing.T) {
	query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)
	table := MetaTestRecord

	t.Run("count error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockAPI := NewMockAPI(ctrl)
		options := qb.NewLimitOffset[int]().SetLimit(10).SetOffset(0)
		expectedErr := errors.New("count failed")

		mockAPI.EXPECT().Count(table, query).Return(int32(0), expectedErr)

		var target []*TestRecord
		res, total, err := SelectWithTotal(mockAPI, table, target, query, options)
		assert.Equal(expectedErr, err)
		assert.Equal(0, total)
		assert.Equal(target, res)
	})

	t.Run("limit zero", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockAPI := NewMockAPI(ctrl)
		options := qb.NewLimitOffset[int]().SetLimit(0).SetOffset(0)

		mockAPI.EXPECT().Count(table, query).Return(int32(15), nil)

		var target []*TestRecord
		res, total, err := SelectWithTotal(mockAPI, table, target, query, options)
		assert.NoError(err)
		assert.Equal(15, total)
		assert.Equal(target, res)
	})

	t.Run("total zero", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockAPI := NewMockAPI(ctrl)
		options := qb.NewLimitOffset[int]().SetLimit(10).SetOffset(0)

		mockAPI.EXPECT().Count(table, query).Return(int32(0), nil)

		var target []*TestRecord
		res, total, err := SelectWithTotal(mockAPI, table, target, query, options)
		assert.NoError(err)
		assert.Equal(0, total)
		assert.Equal(target, res)
	})

	t.Run("select error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockAPI := NewMockAPI(ctrl)
		options := qb.NewLimitOffset[int]().SetLimit(10).SetOffset(0)
		expectedErr := errors.New("select failed")

		mockAPI.EXPECT().Count(table, query).Return(int32(25), nil)
		var target []*TestRecord
		mockAPI.EXPECT().Select(&target, query, options).Return(expectedErr)

		res, total, err := SelectWithTotal(mockAPI, table, target, query, options)
		assert.Equal(expectedErr, err)
		assert.Equal(0, total)
		assert.Equal(target, res)
	})

	t.Run("success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockAPI := NewMockAPI(ctrl)
		options := qb.NewLimitOffset[int]().SetLimit(10).SetOffset(0)

		mockAPI.EXPECT().Count(table, query).Return(int32(25), nil)
		var target []*TestRecord
		mockAPI.EXPECT().Select(&target, query, options).Return(nil)

		res, total, err := SelectWithTotal(mockAPI, table, target, query, options)
		assert.NoError(err)
		assert.Equal(25, total)
		assert.Equal(target, res)
	})
}

func Test_api_Begin(t *testing.T) {
	t.Run("tx already set", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		err := database.Begin()
		assert.NoError(err)
		assert.Equal(mockTx, database.tx)
	})

	t.Run("begin fails", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		client := NewMockClient(ctrl)
		expectedErr := errors.New("begin failed")
		client.EXPECT().Beginx().Return(nil, expectedErr)

		database := &api{
			db:            &transactable{db: client},
			configuration: &InstanceConfig{},
		}

		err := database.Begin()
		assert.EqualError(err, "begin failed")
		assert.Nil(database.tx)
	})

	t.Run("begin succeeds", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		client := NewMockClient(ctrl)
		client.EXPECT().Beginx().Return(&sqlx.Tx{}, nil)

		database := &api{
			db:            &transactable{db: client},
			configuration: &InstanceConfig{},
		}

		err := database.Begin()
		assert.NoError(err)
		assert.NotNil(database.tx)
	})
}

func Test_api_GetTransaction(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	mockTx := transaction.NewMockTransaction(ctrl)
	database := &api{tx: mockTx}

	assert.Equal(mockTx, database.GetTransaction())

	database.tx = nil
	assert.Nil(database.GetTransaction())
}

func Test_api_Rollback(t *testing.T) {
	t.Run("missing transaction", func(t *testing.T) {
		assert := assert1.New(t)
		database := &api{}
		err := database.Rollback()
		assert.Equal(ErrMissingTransaction, err)
	})

	t.Run("rollback error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		expectedErr := errors.New("rollback failed")
		mockTx.EXPECT().Rollback().Return(expectedErr)

		database := &api{tx: mockTx}
		err := database.Rollback()
		assert.Equal(expectedErr, err)
		assert.Nil(database.tx)
	})

	t.Run("rollback success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		mockTx.EXPECT().Rollback().Return(nil)

		database := &api{tx: mockTx}
		err := database.Rollback()
		assert.NoError(err)
		assert.Nil(database.tx)
	})
}

func Test_api_Commit(t *testing.T) {
	t.Run("missing transaction", func(t *testing.T) {
		assert := assert1.New(t)
		database := &api{}
		err := database.Commit()
		assert.Equal(ErrMissingTransaction, err)
	})

	t.Run("commit error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		expectedErr := errors.New("commit failed")
		mockTx.EXPECT().Commit().Return(expectedErr)

		database := &api{tx: mockTx}
		err := database.Commit()
		assert.Equal(expectedErr, err)
		assert.Nil(database.tx)
	})

	t.Run("commit success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		mockTx.EXPECT().Commit().Return(nil)

		database := &api{tx: mockTx}
		err := database.Commit()
		assert.NoError(err)
		assert.Nil(database.tx)
	})
}

func Test_api_CommitOrRollback(t *testing.T) {
	t.Run("missing transaction", func(t *testing.T) {
		assert := assert1.New(t)
		database := &api{}
		err := database.CommitOrRollback(nil)
		assert.Equal(ErrMissingTransaction, err)
	})

	t.Run("commit on nil error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		mockTx.EXPECT().Commit().Return(nil)

		database := &api{
			tx:            mockTx,
			configuration: &InstanceConfig{},
		}
		err := database.CommitOrRollback(nil)
		assert.NoError(err)
		assert.Nil(database.tx)
	})

	t.Run("rollback on error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		mockTx.EXPECT().Rollback().Return(nil)

		database := &api{
			tx:            mockTx,
			configuration: &InstanceConfig{},
		}
		expectedErr := errors.New("operation failed")
		err := database.CommitOrRollback(expectedErr)
		assert.EqualError(err, "operation failed")
		assert.Nil(database.tx)
	})
}

func Test_api_Count_Distinct(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	mockTx := transaction.NewMockTransaction(ctrl)
	database := &api{
		tx:            mockTx,
		configuration: &InstanceConfig{MaxLimit: 100},
	}
	query := qb.SelectDistinct(MetaTestRecord.ID).From(MetaTestRecord)
	expected := generator.Int32()
	mockTx.EXPECT().Select(&countMatcher{count: expected},
		&queryMatcher{t: t, sql: "SELECT DISTINCT COUNT(DISTINCT `test_record`.`id`) as count FROM " +
			"`test_record` AS `test_record`"},
		qb.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
	).Return(nil)
	actual, err := database.Count(MetaTestRecord, query)
	assert.NoError(err)
	assert.Equal(expected, actual)
}

func Test_api_Count_EmptyTarget(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	mockTx := transaction.NewMockTransaction(ctrl)
	database := &api{
		tx:            mockTx,
		configuration: &InstanceConfig{MaxLimit: 100},
	}
	query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)
	mockTx.EXPECT().Select(&emptyCountMatcher{},
		gomock.Any(),
		qb.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
	).Return(nil)
	actual, err := database.Count(MetaTestRecord, query)
	assert.NoError(err)
	assert.Equal(int32(0), actual)
}

func Test_api_Count_SelectError(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	mockTx := transaction.NewMockTransaction(ctrl)
	database := &api{
		tx:            mockTx,
		configuration: &InstanceConfig{MaxLimit: 100},
	}
	query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)
	expectedErr := errors.New("select count failed")
	mockTx.EXPECT().Select(gomock.Any(),
		gomock.Any(),
		qb.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
	).Return(expectedErr)
	actual, err := database.Count(MetaTestRecord, query)
	assert.Equal(expectedErr, err)
	assert.Equal(int32(0), actual)
}

func Test_api_Sum(t *testing.T) {
	query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)

	t.Run("sum success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{
			tx:            mockTx,
			configuration: &InstanceConfig{MaxLimit: 100},
		}

		expectedVal := 42
		mockTx.EXPECT().Select(&sumMatcher{sum: &expectedVal},
			&queryMatcher{t: t, sql: "SELECT SUM(name) as sum FROM `test_record` AS `test_record`"},
			qb.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
		).Return(nil)

		actual, err := database.Sum(MetaTestRecord.Name, query)
		assert.NoError(err)
		assert.Equal(int32(42), actual)
	})

	t.Run("nil sum result", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{
			tx:            mockTx,
			configuration: &InstanceConfig{MaxLimit: 100},
		}

		mockTx.EXPECT().Select(&sumMatcher{sum: nil},
			gomock.Any(),
			qb.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
		).Return(nil)

		actual, err := database.Sum(MetaTestRecord.Name, query)
		assert.NoError(err)
		assert.Equal(int32(0), actual)
	})

	t.Run("empty target", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{
			tx:            mockTx,
			configuration: &InstanceConfig{MaxLimit: 100},
		}

		mockTx.EXPECT().Select(&emptySumMatcher{},
			gomock.Any(),
			qb.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
		).Return(nil)

		actual, err := database.Sum(MetaTestRecord.Name, query)
		assert.NoError(err)
		assert.Equal(int32(0), actual)
	})

	t.Run("select error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{
			tx:            mockTx,
			configuration: &InstanceConfig{MaxLimit: 100},
		}

		expectedErr := errors.New("sum query failed")
		mockTx.EXPECT().Select(gomock.Any(),
			gomock.Any(),
			qb.NewLimitOffset[int]().SetLimit(1).SetOffset(0),
		).Return(expectedErr)

		actual, err := database.Sum(MetaTestRecord.Name, query)
		assert.Equal(expectedErr, err)
		assert.Equal(int32(0), actual)
	})
}

func Test_api_Create(t *testing.T) {
	t.Run("with existing transaction success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{Name: "create_test"}
		mockTx.EXPECT().Create(rec).Return(nil)

		err := database.Create(rec)
		assert.NoError(err)
		assert.Equal(mockTx, database.tx)
	})

	t.Run("with existing transaction error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{Name: "create_test"}
		expectedErr := errors.New("create error")
		mockTx.EXPECT().Create(rec).Return(expectedErr)

		err := database.Create(rec)
		assert.Equal(expectedErr, err)
		assert.Equal(mockTx, database.tx)
	})

	t.Run("with no transaction begin error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		client := NewMockClient(ctrl)
		expectedErr := errors.New("begin error")
		client.EXPECT().Beginx().Return(nil, expectedErr)

		database := &api{
			db:            &transactable{db: client},
			configuration: &InstanceConfig{},
		}

		rec := &TestRecord{Name: "create_test"}
		err := database.Create(rec)
		assert.EqualError(err, "begin error")
	})
}

func Test_api_Read(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{ID: "read-id"}
		pk := rec.PrimaryKey()
		mockTx.EXPECT().Read(rec, pk).Return(nil)

		err := database.Read(rec, pk)
		assert.NoError(err)
	})

	t.Run("error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{ID: "read-id"}
		pk := rec.PrimaryKey()
		expectedErr := errors.New("read error")
		mockTx.EXPECT().Read(rec, pk).Return(expectedErr)

		err := database.Read(rec, pk)
		assert.Equal(expectedErr, err)
	})
}

func Test_api_ReadOneWhere(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{}
		cond := qb.FieldComparison(MetaTestRecord.ID, qb.Equal, "123")
		mockTx.EXPECT().ReadOneWhere(rec, cond).Return(nil)

		err := database.ReadOneWhere(rec, cond)
		assert.NoError(err)
	})

	t.Run("error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{}
		cond := qb.FieldComparison(MetaTestRecord.ID, qb.Equal, "123")
		expectedErr := errors.New("read one error")
		mockTx.EXPECT().ReadOneWhere(rec, cond).Return(expectedErr)

		err := database.ReadOneWhere(rec, cond)
		assert.Equal(expectedErr, err)
	})
}

func Test_api_Select(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{
			tx:            mockTx,
			configuration: &InstanceConfig{MaxLimit: 100},
		}

		var target []*TestRecord
		query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)
		options := qb.NewLimitOffset[int]().SetLimit(50).SetOffset(10)
		mockTx.EXPECT().Select(&target, query, options).Return(nil)

		err := database.Select(&target, query, options)
		assert.NoError(err)
	})

	t.Run("error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{
			tx:            mockTx,
			configuration: &InstanceConfig{MaxLimit: 100},
		}

		var target []*TestRecord
		query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)
		options := qb.NewLimitOffset[int]().SetLimit(50).SetOffset(10)
		expectedErr := errors.New("select error")
		mockTx.EXPECT().Select(&target, query, options).Return(expectedErr)

		err := database.Select(&target, query, options)
		assert.Equal(expectedErr, err)
	})
}

func Test_api_SelectOne(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{}
		query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)
		mockTx.EXPECT().SelectOne(rec, query).Return(nil)

		err := database.SelectOne(rec, query)
		assert.NoError(err)
	})

	t.Run("error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{}
		query := qb.Select(MetaTestRecord.ID).From(MetaTestRecord)
		expectedErr := errors.New("select one error")
		mockTx.EXPECT().SelectOne(rec, query).Return(expectedErr)

		err := database.SelectOne(rec, query)
		assert.Equal(expectedErr, err)
	})
}

func Test_api_ListWhere(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{
			tx:            mockTx,
			configuration: &InstanceConfig{MaxLimit: 100},
		}

		meta := &TestRecord{}
		var target []*TestRecord
		cond := qb.FieldComparison(MetaTestRecord.Name, qb.Equal, "test")
		options := qb.NewLimitOffset[int]().SetLimit(50).SetOffset(10)
		mockTx.EXPECT().ListWhere(meta, &target, cond, options).Return(nil)

		err := database.ListWhere(meta, &target, cond, options)
		assert.NoError(err)
	})

	t.Run("error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{
			tx:            mockTx,
			configuration: &InstanceConfig{MaxLimit: 100},
		}

		meta := &TestRecord{}
		var target []*TestRecord
		cond := qb.FieldComparison(MetaTestRecord.Name, qb.Equal, "test")
		options := qb.NewLimitOffset[int]().SetLimit(50).SetOffset(10)
		expectedErr := errors.New("list where error")
		mockTx.EXPECT().ListWhere(meta, &target, cond, options).Return(expectedErr)

		err := database.ListWhere(meta, &target, cond, options)
		assert.Equal(expectedErr, err)
	})
}

func Test_api_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{ID: "update-id", Name: "updated"}
		mockTx.EXPECT().Update(rec).Return(nil)

		err := database.Update(rec)
		assert.NoError(err)
	})

	t.Run("error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{ID: "update-id", Name: "updated"}
		expectedErr := errors.New("update error")
		mockTx.EXPECT().Update(rec).Return(expectedErr)

		err := database.Update(rec)
		assert.Equal(expectedErr, err)
	})
}

func Test_api_UpdateWhere(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{}
		cond := qb.FieldComparison(MetaTestRecord.ID, qb.Equal, "123")
		fv := qb.FieldValue{Field: MetaTestRecord.Name, Value: "new_name"}
		mockTx.EXPECT().UpdateWhere(rec, cond, fv).Return(int64(4), nil)

		total, err := database.UpdateWhere(rec, cond, fv)
		assert.NoError(err)
		assert.Equal(int64(4), total)
	})

	t.Run("error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{}
		cond := qb.FieldComparison(MetaTestRecord.ID, qb.Equal, "123")
		fv := qb.FieldValue{Field: MetaTestRecord.Name, Value: "new_name"}
		expectedErr := errors.New("update where error")
		mockTx.EXPECT().UpdateWhere(rec, cond, fv).Return(int64(0), expectedErr)

		total, err := database.UpdateWhere(rec, cond, fv)
		assert.Equal(expectedErr, err)
		assert.Equal(int64(0), total)
	})
}

func Test_api_UpdateIgnoreWhere(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{}
		cond := qb.FieldComparison(MetaTestRecord.ID, qb.Equal, "123")
		fv := qb.FieldValue{Field: MetaTestRecord.Name, Value: "new_name"}
		mockTx.EXPECT().UpdateIgnoreWhere(rec, cond, fv).Return(int64(2), nil)

		total, err := database.UpdateIgnoreWhere(rec, cond, fv)
		assert.NoError(err)
		assert.Equal(int64(2), total)
	})

	t.Run("error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{}
		cond := qb.FieldComparison(MetaTestRecord.ID, qb.Equal, "123")
		fv := qb.FieldValue{Field: MetaTestRecord.Name, Value: "new_name"}
		expectedErr := errors.New("update ignore where error")
		mockTx.EXPECT().UpdateIgnoreWhere(rec, cond, fv).Return(int64(0), expectedErr)

		total, err := database.UpdateIgnoreWhere(rec, cond, fv)
		assert.Equal(expectedErr, err)
		assert.Equal(int64(0), total)
	})
}

func Test_api_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{ID: "delete-id"}
		mockTx.EXPECT().Delete(rec).Return(nil)

		err := database.Delete(rec)
		assert.NoError(err)
	})

	t.Run("error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{ID: "delete-id"}
		expectedErr := errors.New("delete error")
		mockTx.EXPECT().Delete(rec).Return(expectedErr)

		err := database.Delete(rec)
		assert.Equal(expectedErr, err)
	})
}

func Test_api_DeleteWhere(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{}
		cond := qb.FieldComparison(MetaTestRecord.ID, qb.Equal, "123")
		mockTx.EXPECT().DeleteWhere(rec, cond).Return(nil)

		err := database.DeleteWhere(rec, cond)
		assert.NoError(err)
	})

	t.Run("error", func(t *testing.T) {
		assert := assert1.New(t)
		ctrl := gomock.NewController(t)
		mockTx := transaction.NewMockTransaction(ctrl)
		database := &api{tx: mockTx}

		rec := &TestRecord{}
		cond := qb.FieldComparison(MetaTestRecord.ID, qb.Equal, "123")
		expectedErr := errors.New("delete where error")
		mockTx.EXPECT().DeleteWhere(rec, cond).Return(expectedErr)

		err := database.DeleteWhere(rec, cond)
		assert.Equal(expectedErr, err)
	})
}
