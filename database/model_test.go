package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
)

func TestNewPrimaryKey(t *testing.T) {
	assert := assert1.New(t)

	pk := NewPrimaryKey(1)
	assert.Equal(1, pk.Value())

	pk = NewPrimaryKey("foo")
	assert.Equal("foo", pk.Value())
}

func TestNewListOptions(t *testing.T) {
	assert := assert1.New(t)

	data := []struct {
		Expected *ListOptions
		Limit    uint
		Offset   uint
	}{
		{&ListOptions{Limit: 0, Offset: 0}, 0, 0},
		{&ListOptions{Limit: DefaultMaxLimit, Offset: 10}, DefaultMaxLimit, 10},
		{&ListOptions{Limit: 5, Offset: 5}, 5, 5},
	}
	for _, test := range data {
		actual := NewListOptions(test.Limit, test.Offset)
		assert.Equal(test.Expected, actual)
	}
}

func TestBadConnection(t *testing.T) {
	assert := assert1.New(t)
	_, err := connect("mysql", "baduser:badpassword@tcp(localhost:3306)/foo", log.NewStackLogger())
	assert.EqualError(err, NewDatabaseConnectionError(err).Error())
}

func TestBadRecordImplementation(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	record := &DetailsTestRecord{}

	err := spec.DB.Create(record)
	assert.IsType(&SQLExecutionError{}, err)
	assert.IsType(&SQLExecutionError{}, spec.DB.Read(record, NewPrimaryKey("foo")))
	options := NewListOptions(10, 0)
	assert.IsType(&SQLExecutionError{}, spec.DB.List(record, []DetailsTestRecord{}, options))
	assert.IsType(&SQLExecutionError{}, spec.DB.Update(record))
	assert.IsType(&SQLExecutionError{}, spec.DB.Delete(record))
}

func TestCreate(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	name := generator.Name()
	record := &TestRecord{Name: name}
	assert.NoError(spec.DB.Create(record))
	assert.Equal(name, record.Name)
	assert.Equal("tst", record.ID[:3])
	assert.Equal(record.PrimaryKey().Value(), record.ID)
}

func TestCreateDuplicate(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	name := generator.Name()
	record := &TestRecord{Name: name}
	assert.NoError(spec.DB.Create(record))
	assert.Equal(name, record.Name)
	assert.Equal("tst", record.ID[:3])

	record = &TestRecord{Name: name}
	err := spec.DB.Create(record)
	assert.IsType(&UniqueConstraintError{}, err)
}

func TestRead(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	expected := &TestRecord{Name: generator.Name()}
	assert.NoError(spec.DB.Create(expected))

	actual := &TestRecord{}
	assert.NoError(spec.DB.Read(actual, expected.PrimaryKey()))
	assert.Equal(expected, actual)
}

func TestReadOneWhere(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	expected := &TestRecord{Name: generator.Name()}
	assert.NoError(spec.DB.Create(expected))

	actual := &TestRecord{}
	assert.NoError(spec.DB.ReadOneWhere(actual, TestMeta.Name.Equal(expected.Name)))
	assert.Equal(expected, actual)
}

func TestList(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	records := make([]TestRecord, 5)
	for i := range records {
		record := &TestRecord{Name: fmt.Sprintf("Test %s", strconv.Itoa(i))}
		assert.NoError(spec.DB.Create(record))
		records[i] = *record
	}

	lookup := &[]TestRecord{}
	options := NewListOptions(3, 0)
	assert.NoError(spec.DB.List(&TestRecord{}, lookup, options))
	assert.Equal(3, len(*lookup))
}

func TestList_QueryLimit(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()
	conf := spec.DB.Configuration.(*specification)
	conf.QueryLimit = 2

	records := make([]TestRecord, 5)
	for i := range records {
		record := &TestRecord{Name: fmt.Sprintf("TestList_QueryLimit %s", strconv.Itoa(i))}
		assert.NoError(spec.DB.Create(record))
		records[i] = *record
	}

	lookup := &[]TestRecord{}
	options := NewListOptions(conf.QueryLimit+5, 0)
	assert.NoError(spec.DB.List(&TestRecord{}, lookup, options))
	assert.Equal(int(conf.QueryLimit), len(*lookup))

	lookup = &[]TestRecord{}
	options = NewListOptions(conf.QueryLimit-1, 0)
	assert.NoError(spec.DB.List(&TestRecord{}, lookup, options))
	assert.Equal(int(conf.QueryLimit-1), len(*lookup))
}

func TestWhere(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	records := make([]*TestRecord, 5)
	for i := range records {
		record := &TestRecord{Name: generator.Name()}
		assert.NoError(spec.DB.Create(record))
		records[i] = record
	}

	where := TestMeta.Name.Equal(records[0].Name)
	var actual []*TestRecord
	err := spec.DB.ListWhere(&TestRecord{}, &actual, where, nil)
	if assert.NoError(err) {
		assert.Equal(records[0], actual[0])
		assert.Equal(1, len(actual))
	}
	where = TestMeta.Name.Equal(records[0])
	err = spec.DB.ListWhere(&TestRecord{}, &actual, where, nil)
	assert.Error(err)
}

func TestWhereTx(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	tx := spec.DB.MustBegin()
	defer func() { assert.NoError(tx.Commit()) }()
	records := make([]*TestRecord, 5)
	for i := range records {
		record := &TestRecord{Name: generator.Name()}
		assert.NoError(spec.DB.CreateTx(record, tx))
		records[i] = record
	}

	where := TestMeta.Name.Equal(records[0].Name)
	var actual []*TestRecord
	err := spec.DB.ListWhere(&TestRecord{}, &actual, where, nil)
	if assert.NoError(err) {
		assert.Equal(0, len(actual))
	}

	err = spec.DB.ListWhereTx(tx, &TestRecord{}, &actual, where, nil)
	if assert.NoError(err) {
		assert.Equal(records[0], actual[0])
		assert.Equal(1, len(actual))
	}
	where = TestMeta.Name.Equal(records[0])
	err = spec.DB.ListWhereTx(tx, &TestRecord{}, &actual, where, nil)
	assert.Error(err)
}

func TestWhereIn(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	records := make([]*TestRecord, 5)
	for i := range records {
		record := &TestRecord{Name: generator.Name()}
		assert.NoError(spec.DB.Create(record))
		records[i] = record
		time.Sleep(time.Second)
	}

	where := TestMeta.Name.In(records[0].Name, records[1].Name)

	var actual []*TestRecord
	err := spec.DB.ListWhere(&TestRecord{}, &actual, where, nil)
	if assert.NoError(err) {
		assert.Equal(records[0], actual[0])
		assert.Equal(records[1], actual[1])
		assert.Equal(2, len(actual))
	}
	where = TestMeta.Name.In(records[0])
	assert.Error(spec.DB.ListWhere(&TestRecord{}, &actual, where, nil))

	where = TestMeta.Name.In(records[0].Name, records[1].Name).And(TestMeta.ID.Equal(records[0].ID))
	actual = []*TestRecord{}
	err = spec.DB.ListWhere(&TestRecord{}, &actual, where, nil)
	if assert.NoError(err) {
		assert.Equal(records[0], actual[0])
		assert.Equal(1, len(actual))
	}

	where = TestMeta.ID.Equal(records[1].ID).And(TestMeta.Name.In(records[0].Name, records[1].Name))
	actual = []*TestRecord{}
	err = spec.DB.ListWhere(&TestRecord{}, &actual, where, nil)
	if assert.NoError(err) {
		assert.Equal(records[1], actual[0])
		assert.Equal(1, len(actual))
	}
}

func TestSelectList(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	records := make([]*TestRecord, 5)
	for i := range records {
		name := fmt.Sprintf("%d%s", i, generator.Name())
		record := &TestRecord{Name: name}
		assert.NoError(spec.DB.Create(record))
		records[i] = record
	}

	query := qb.Select(TestMeta.AllColumns()).From(TestMeta).
		Where(TestMeta.Name.In(records[0].Name, records[1].Name)).
		OrderBy(TestMeta.Name, qb.Ascending)
	var actual []*TestRecord
	err := spec.DB.SelectList(&actual, query, nil)
	if assert.NoError(err) {
		assert.Equal(2, len(actual))
		assert.Equal(records[0], actual[0])
		assert.Equal(records[1], actual[1])
	}

	actual = nil
	options := NewListOptions(3, 1)
	err = spec.DB.SelectList(&actual, query, options)
	if assert.NoError(err) {
		assert.Equal(1, len(actual))
		assert.Equal(records[1], actual[0])
	}
}

func TestSelectTxList(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	tx := spec.DB.MustBegin()
	defer func() { assert.NoError(tx.Commit()) }()
	records := make([]*TestRecord, 5)
	for i := range records {
		name := fmt.Sprintf("%d%s", i, generator.Name())
		record := &TestRecord{Name: name}
		assert.NoError(spec.DB.CreateTx(record, tx))
		records[i] = record
	}

	query := qb.Select(TestMeta.AllColumns()).From(TestMeta).
		Where(TestMeta.Name.In(records[0].Name, records[1].Name)).
		OrderBy(TestMeta.Name, qb.Ascending)
	var actual []*TestRecord
	err := spec.DB.SelectList(&actual, query, nil)
	if assert.NoError(err) {
		assert.Equal(0, len(actual))
	}

	actual = nil
	err = spec.DB.SelectListTx(tx, &actual, query, nil)
	if assert.NoError(err) {
		assert.Equal(2, len(actual))
		assert.Equal(records[0], actual[0])
		assert.Equal(records[1], actual[1])
	}

	actual = nil
	options := NewListOptions(3, 1)
	err = spec.DB.SelectListTx(tx, &actual, query, options)
	if assert.NoError(err) {
		assert.Equal(1, len(actual))
		assert.Equal(records[1], actual[0])
	}
}

func TestDelete(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	expected := &TestRecord{Name: generator.Name()}
	assert.NoError(spec.DB.Create(expected))

	actual := &TestRecord{}
	assert.NoError(spec.DB.Read(actual, expected.PrimaryKey()))
	assert.Equal(expected, actual)

	assert.NoError(spec.DB.Delete(actual))
	assert.IsType(&NotFoundError{}, spec.DB.Read(actual, expected.PrimaryKey()))
	assert.NoError(spec.DB.Delete(actual))
}

func TestDeleteTx(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	expected := &TestRecord{Name: generator.Name()}
	assert.NoError(spec.DB.Create(expected))

	actual := &TestRecord{}
	assert.NoError(spec.DB.Read(actual, expected.PrimaryKey()))
	assert.Equal(expected, actual)

	tx := spec.DB.MustBegin()
	assert.NoError(spec.DB.DeleteTx(actual, tx))
	log.Error(tx.Rollback())
	assert.NoError(spec.DB.Read(actual, expected.PrimaryKey()))

	tx = spec.DB.MustBegin()
	assert.NoError(spec.DB.DeleteTx(actual, tx))
	log.Error(tx.Commit())
	assert.IsType(&NotFoundError{}, spec.DB.Read(actual, expected.PrimaryKey()))
}

func TestDeleteWhere(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	expected := &TestRecord{Name: generator.Name()}
	assert.NoError(spec.DB.Create(expected))

	actual := &TestRecord{}
	assert.NoError(spec.DB.Read(actual, expected.PrimaryKey()))
	assert.Equal(expected, actual)

	assert.NoError(spec.DB.DeleteWhere(actual, TestMeta.Name.Equal(expected.Name)))
	assert.IsType(&NotFoundError{}, spec.DB.Read(actual, expected.PrimaryKey()))
	assert.NoError(spec.DB.DeleteWhere(actual, TestMeta.Name.Equal(expected.Name)))
}

func TestDeleteWhereTx(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	expected := &TestRecord{Name: "DeleteWhereMe"}
	assert.NoError(spec.DB.Create(expected))

	actual := &TestRecord{}
	assert.NoError(spec.DB.Read(actual, expected.PrimaryKey()))
	assert.Equal(expected, actual)

	tx := spec.DB.MustBegin()
	assert.NoError(spec.DB.DeleteWhereTx(actual, tx, TestMeta.Name.Equal(expected.Name)))
	assert.NoError(tx.Rollback())
	assert.NoError(spec.DB.Read(actual, expected.PrimaryKey()))

	tx = spec.DB.MustBegin()
	assert.NoError(spec.DB.DeleteWhereTx(actual, tx, TestMeta.Name.Equal(expected.Name)))
	assert.NoError(tx.Commit())
	assert.IsType(&NotFoundError{}, spec.DB.Read(actual, expected.PrimaryKey()))
}

func TestUpdate(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	expected := &TestRecord{Name: "Update Me"}
	assert.NoError(spec.DB.Create(expected))

	actual := &TestRecord{}
	assert.NoError(spec.DB.Read(actual, expected.PrimaryKey()))
	assert.Equal(expected, actual)

	expected.Name = "Updated"
	assert.NoError(spec.DB.Update(expected))
	assert.NoError(spec.DB.Read(actual, expected.PrimaryKey()))
	assert.Equal(expected, actual)
}

func TestUpsertTx(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	recordID := generator.ID("tst")
	record := &TestRecord{
		ID:        recordID,
		Name:      generator.Name(),
		CreatedOn: time.Now(),
		UpdatedOn: time.Now(),
	}
	tx := spec.DB.MustBegin()
	updatedOn := record.UpdatedOn
	assert.NoError(spec.DB.UpsertTx(record, tx))
	assert.Equal(recordID, record.ID)
	assert.NotEqual(updatedOn, record.UpdatedOn)

	updatedOn = record.UpdatedOn
	err := spec.DB.UpsertTx(record, tx)
	assert.NoError(err)
	assert.Equal(updatedOn, record.UpdatedOn)

	time.Sleep(2 * time.Second)
	expected := generator.Name()
	assert.NotEqual(expected, record.Name)
	record.Name = expected
	assert.NoError(spec.DB.UpsertTx(record, tx))
	assert.Equal(updatedOn, record.UpdatedOn)
	assert.Equal(expected, record.Name)
	assert.NoError(tx.Rollback())
}

func TestUpdateWhere(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	expected := &TestRecord{Name: "TestUpdateWhere"}
	assert.NoError(spec.DB.Create(expected))

	actual := &TestRecord{}
	assert.NoError(spec.DB.Read(actual, expected.PrimaryKey()))
	assert.Equal(expected, actual)

	affected, err := spec.DB.UpdateWhere(expected, TestMeta.ID.Equal(expected.ID),
		qb.FieldValue{Field: TestMeta.Name, Value: "TestUpdateWhere updated"},
		qb.FieldValue{Field: TestMeta.Place, Value: "Place"})
	assert.NoError(err)
	assert.Equal(int64(1), affected)

	assert.NoError(spec.DB.Read(actual, expected.PrimaryKey()))
	expected.Name = "TestUpdateWhere updated"
	expected.Place = sql.NullString{
		String: "Place",
		Valid:  true,
	}

	assert.Equal(expected, actual)
}

func TestUpdateWhereTx(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()
	tx := spec.DB.MustBegin()
	t.Cleanup(func() {
		_ = tx.Rollback()
	})

	expected := &TestRecord{Name: "TestUpdateWhereTx"}
	assert.NoError(spec.DB.CreateTx(expected, tx))

	actual := &TestRecord{}
	assert.NoError(spec.DB.ReadTx(actual, expected.PrimaryKey(), tx))
	assert.Equal(expected, actual)

	affected, err := spec.DB.UpdateWhereTx(expected, tx, TestMeta.ID.Equal(expected.ID),
		qb.FieldValue{Field: TestMeta.Name, Value: "TestUpdateWhereTx updated"},
		qb.FieldValue{Field: TestMeta.Place, Value: "Place"})
	assert.NoError(err)
	assert.Equal(int64(1), affected)

	assert.NoError(spec.DB.ReadTx(actual, expected.PrimaryKey(), tx))
	expected.Name = "TestUpdateWhereTx updated"
	expected.Place = sql.NullString{
		String: "Place",
		Valid:  true,
	}

	assert.Equal(expected, actual)
}
func TestTransaction(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	name := generator.Name()
	record := &TestRecord{Name: name}
	tx, err := spec.DB.Beginx()
	assert.NoError(err)

	assert.NoError(spec.DB.CreateTx(record, tx))
	assert.Equal(name, record.Name)
	assert.Equal("tst", record.ID[:3])
	assert.Equal(record.PrimaryKey().Value(), record.ID)

	actual := &TestRecord{}
	err = spec.DB.ReadTx(actual, record.PrimaryKey(), tx)
	assert.NoError(err)
	assert.NoError(tx.Rollback())

	err = spec.DB.Read(actual, record.PrimaryKey())
	assert.EqualError(err, NewNotFoundError().Error())
}

type initDupe struct {
	base string
}

func (id *initDupe) id() string {
	defer func() { id.base = generator.ID("nodupe") }()
	return id.base
}

func TestDuplicateID(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	record := NewTestDuper()
	assert.NoError(spec.DB.Create(record))
	assert.Equal("dup", record.ID[:3])

	record2 := NewTestDuper()
	assert.NoError(spec.DB.Create(record2))
	assert.Equal("dup", record2.ID[:3])
	assert.NotEqual(record2.ID, record.ID)

	record3 := NewTestDuper()
	record3.intializer = func() string { return record.ID }
	assert.Error(spec.DB.Create(record3))

	record3 = NewTestDuper()
	initDuper := &initDupe{base: record.ID}
	initDuper.id()
	record3.intializer = func() string { return initDuper.id() }

	assert.NoError(spec.DB.Create(record3))
	assert.Equal("nodupe", record3.ID[:6])
	assert.NotEqual(record3.ID, record.ID)
}

func TestIsNotFoundError(t *testing.T) {
	tcs := []struct {
		name string
		err  error
		res  bool
	}{
		{
			name: "is not found",
			err:  &NotFoundError{},
			res:  true,
		},
		{
			name: "is other error",
			err:  errors.New("other error"),
			res:  false,
		}, {
			name: "is nil",
			err:  nil,
			res:  false,
		},
	}

	for _, tc := range tcs {
		assert := assert1.New(t)

		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(tc.res, IsNotFoundError(tc.err))
		})
	}
}

func TestTableExists(t *testing.T) {
	spec := newSpecification()

	tcs := []struct {
		name       string
		schemaName string
		tableName  string
		res        bool
	}{
		{
			name:       "table exists",
			schemaName: "test_db",
			tableName:  "test_duper",
			res:        true,
		},
		{
			name:       "table not exists",
			schemaName: "test_db",
			tableName:  "no_such_table",
			res:        false,
		},
		{
			name:       "schema not exists",
			schemaName: "no_such_schema",
			tableName:  "test_duper",
			res:        false,
		},
		{
			name:       "table and schema not exists",
			schemaName: "no_such_schema",
			tableName:  "no_such_table",
			res:        false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert1.New(t)
			ok, err := TableExists(spec.DB, tc.schemaName, tc.tableName)
			assert.NoError(err)
			assert.Equal(tc.res, ok)
		})
	}

}

func TestDatabaseLock(t *testing.T) {
	assert := assert1.New(t)
	spec1 := newSpecification()
	spec2 := newSpecification()

	lockName := uuid.New().String()

	// session 1 should acquire lock
	ok, err := AcquireDatabaseLock(spec1.DB, lockName, 5)
	assert.NoError(err)
	assert.True(ok)

	// session 2 should fail to acquire same lock
	ok, err = AcquireDatabaseLock(spec2.DB, lockName, 5)
	assert.NoError(err)
	assert.False(ok)

	// session 2 should acquire different lock
	ok, err = AcquireDatabaseLock(spec2.DB, uuid.New().String(), 5)
	assert.NoError(err)
	assert.True(ok)

	// session 1 should release lock
	err = ReleaseDatabaseLock(spec1.DB, lockName)
	assert.NoError(err)

	// session 2 should acquire lock
	ok, err = AcquireDatabaseLock(spec2.DB, lockName, 5)
	assert.NoError(err)
	assert.True(ok)
}

func TestSelect(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()

	tx := spec.DB.MustBegin()
	defer func() { assert.NoError(tx.Commit()) }()
	records := make([]*TestRecord, 5)
	for i := range records {
		name := fmt.Sprintf("%d%s", i, generator.Name())
		record := &TestRecord{Name: name}
		assert.NoError(spec.DB.CreateTx(record, tx))
		records[i] = record
	}

	query := qb.Select(TestMeta.AllColumns()).From(TestMeta).
		Where(TestMeta.Name.In(records[0].Name, records[1].Name)).
		OrderBy(TestMeta.Name, qb.Ascending)
	var actual []*TestRecord
	err := spec.DB.Select(&actual, query)
	if assert.NoError(err) {
		assert.Equal(0, len(actual))
	}

	actual = nil
	err = spec.DB.SelectTx(tx, &actual, query)
	if assert.NoError(err) {
		assert.Equal(2, len(actual))
	}
}

func TestSelect_LimitEnforced(t *testing.T) {
	assert := assert1.New(t)
	spec := newSpecification()
	conf := spec.DB.Configuration.(*specification)
	conf.QueryLimit = 3

	records := make([]*TestRecord, 5)
	for i := range records {
		name := fmt.Sprintf("%d%s", i, generator.Name())
		record := &TestRecord{Name: name}
		assert.NoError(spec.DB.Create(record))
		records[i] = record
	}

	query := qb.Select(TestMeta.AllColumns()).From(TestMeta).
		OrderBy(TestMeta.Name, qb.Ascending)
	var actual []*TestRecord
	err := spec.DB.Select(&actual, query)
	if assert.NoError(err) {
		assert.Equal(int(conf.QueryLimit), len(actual))
	}
}

func Test_obfuscateConnect(t *testing.T) {
	tests := []struct {
		name     string
		arg      string
		expected string
	}{
		{
			name:     "empty",
			arg:      "",
			expected: "",
		},
		{
			name:     "no credentials",
			arg:      "foo.com",
			expected: "foo.com",
		},
		{
			name:     "credentials",
			arg:      "user:password@foo.com",
			expected: "*************@foo.com",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := obfuscateConnection(test.arg)
			if actual != test.expected {
				t.Errorf("obfuscateConnection(\"%s\") = \"%s\", expected \"%s\"",
					test.arg, actual, test.expected)
			}
		})
	}
}
