package qb

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
)

func TestInsertQuery(t *testing.T) {
	assert := assert1.New(t)
	pID := generator.TestID()
	name := generator.String(5)
	aID := generator.TestID()
	query := Insert(Person.ID, Person.Name, Person.AddressID).Values(pID, name, aID)
	sql, values, err := query.SQL()
	assert.NoError(err)
	if assert.Equal(3, len(values)) {
		values[0] = pID
		values[1] = name
		values[2] = aID
	}
	assert.Equal("INSERT INTO `person` (`person`.`id`, `person`.`name`, `person`.`address_id`) VALUES (?, ?, ?)", sql)
}

func TestInsertQueryMulti(t *testing.T) {
	assert := assert1.New(t)
	query := Insert(Person.ID, Person.Name, Person.AddressID)
	for i := 0; i < 10; i++ {
		query.Values(i, i, i)
	}
	sql, values, err := query.SQL()
	assert.NoError(err)
	assert.Equal(30, len(values))
	for i := 0; i < 10; i++ {
		j := i * 3
		assert.Equal(i, values[j])
		assert.Equal(i, values[j+1])
		assert.Equal(i, values[j+2])
	}
	assert.Equal("INSERT INTO `person` (`person`.`id`, `person`.`name`, `person`.`address_id`) VALUES "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?)", sql)
}

func TestInsertMismatchColumnValuesCount(t *testing.T) {
	assert := assert1.New(t)
	query := Insert(Person.ID, Person.Name, Person.AddressID).Values(1, 2)
	assert.EqualError(query.err, "insert field/value count mismatch")
	_, _, err := query.SQL()
	assert.EqualError(err, "insert field/value count mismatch")
}

func TestInsertNoColumns(t *testing.T) {
	assert := assert1.New(t)
	query := Insert()
	_, _, err := query.SQL()
	assert.EqualError(err, "no columns specified for insert")
}

func TestInsertDifferentTablesError(t *testing.T) {
	assert := assert1.New(t)
	query := Insert(Person.ID, Address.ID)
	_, _, err := query.SQL()
	assert.EqualError(err, "insert columns must be from the same table")
}

func TestInsertDifferentDuplicateTablesError(t *testing.T) {
	assert := assert1.New(t)
	query := Insert(Person.ID).OnDuplicate(Address.WriteColumns())
	_, _, err := query.SQL()
	assert.EqualError(err, "duplicate columns must be from the same table")
}

func TestInsertQueryOnDuplicate(t *testing.T) {
	assert := assert1.New(t)
	pID := generator.TestID()
	name := generator.String(5)
	aID := generator.TestID()
	query := Insert(Person.ID, Person.Name, Person.AddressID).Values(pID, name, aID).OnDuplicate(Person.WriteColumns())
	sql, values, err := query.SQL()
	assert.NoError(err)
	if assert.Equal(3, len(values)) {
		values[0] = pID
		values[1] = name
		values[2] = aID
	}
	assert.Equal("INSERT INTO `person` (`person`.`id`, `person`.`name`, `person`.`address_id`) VALUES "+
		"(?, ?, ?) "+
		"ON DUPLICATE KEY UPDATE "+
		"`person`.`id` = VALUES(`person`.`id`), "+
		"`person`.`name` = VALUES(`person`.`name`), "+
		"`person`.`address_id` = VALUES(`person`.`address_id`), "+
		"`person`.`age` = VALUES(`person`.`age`)", sql)
}

func TestInsertQueryMultiOnDuplicate(t *testing.T) {
	assert := assert1.New(t)
	query := Insert(Person.ID, Person.Name, Person.AddressID).OnDuplicate(Person.WriteColumns())
	for i := 0; i < 10; i++ {
		query.Values(i, i, i)
	}
	sql, _, err := query.SQL()
	assert.NoError(err)
	assert.Equal("INSERT INTO `person` (`person`.`id`, `person`.`name`, `person`.`address_id`) VALUES "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?), "+
		"(?, ?, ?) "+
		"ON DUPLICATE KEY UPDATE "+
		"`person`.`id` = VALUES(`person`.`id`), "+
		"`person`.`name` = VALUES(`person`.`name`), "+
		"`person`.`address_id` = VALUES(`person`.`address_id`), "+
		"`person`.`age` = VALUES(`person`.`age`)", sql)
	sql, err = query.ParameterizedSQL()
	assert.NoError(err)
	assert.Equal("INSERT INTO `person` (`person`.`id`, `person`.`name`, `person`.`address_id`) VALUES "+
		"(:id, :name, :address_id) "+
		"ON DUPLICATE KEY UPDATE "+
		"`person`.`id` = VALUES(`person`.`id`), "+
		"`person`.`name` = VALUES(`person`.`name`), "+
		"`person`.`address_id` = VALUES(`person`.`address_id`), "+
		"`person`.`age` = VALUES(`person`.`age`)", sql)
}
