package database

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/generator"
	"github.com/beaconsoftwarellc/gadget/log"
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
		{&ListOptions{Limit: 1, Offset: 0}, 0, 0},
		{&ListOptions{Limit: MaxLimit, Offset: 10}, MaxLimit + 1, 10},
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
	assert.EqualError(err, NewDatabaseConnectionError().Error())
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
	err := spec.DB.ListWhere(&TestRecord{}, &actual, where)
	if assert.NoError(err) {
		assert.Equal(records[0], actual[0])
		assert.Equal(1, len(actual))
	}
	where = TestMeta.Name.Equal(records[0])
	err = spec.DB.ListWhere(&TestRecord{}, &actual, where)
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
	err := spec.DB.ListWhere(&TestRecord{}, &actual, where)
	if assert.NoError(err) {
		assert.Equal(0, len(actual))
	}

	err = spec.DB.ListWhereTx(tx, &TestRecord{}, &actual, where)
	if assert.NoError(err) {
		assert.Equal(records[0], actual[0])
		assert.Equal(1, len(actual))
	}
	where = TestMeta.Name.Equal(records[0])
	err = spec.DB.ListWhereTx(tx, &TestRecord{}, &actual, where)
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
	err := spec.DB.ListWhere(&TestRecord{}, &actual, where)
	if assert.NoError(err) {
		assert.Equal(records[0], actual[0])
		assert.Equal(records[1], actual[1])
		assert.Equal(2, len(actual))
	}
	where = TestMeta.Name.In(records[0])
	assert.Error(spec.DB.ListWhere(&TestRecord{}, &actual, where))

	where = TestMeta.Name.In(records[0].Name, records[1].Name).And(TestMeta.ID.Equal(records[0].ID))
	actual = []*TestRecord{}
	err = spec.DB.ListWhere(&TestRecord{}, &actual, where)
	if assert.NoError(err) {
		assert.Equal(records[0], actual[0])
		assert.Equal(1, len(actual))
	}

	where = TestMeta.ID.Equal(records[1].ID).And(TestMeta.Name.In(records[0].Name, records[1].Name))
	actual = []*TestRecord{}
	err = spec.DB.ListWhere(&TestRecord{}, &actual, where)
	if assert.NoError(err) {
		assert.Equal(records[1], actual[0])
		assert.Equal(1, len(actual))
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
	i    int
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
