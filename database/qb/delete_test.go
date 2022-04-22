package qb

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
)

func TestDeleteQuery(t *testing.T) {
	assert := assert1.New(t)
	expectedID := generator.TestID()
	query := Delete(Person)
	query.Where(Person.ID.Equal(expectedID))
	actual, values, err := query.SQL()
	assert.NoError(err)
	if assert.Equal(len(values), 1) {
		assert.Equal(expectedID, values[0])
	}
	assert.Equal("DELETE FROM `person` WHERE `person`.`id` = ?", actual)
}

func TestDeleteQueryInnerJoin(t *testing.T) {
	assert := assert1.New(t)
	query := Delete(Person, Address).From(Person)
	query.InnerJoin(Address).On(Address.ID, Equal, Person.AddressID)
	query.Where(Address.Country.NotEqual("US"))
	actual, values, err := query.SQL()
	assert.NoError(err)
	if assert.Equal(1, len(values)) {
		assert.Equal("US", values[0])
	}
	assert.Equal("DELETE `person`, `address` "+
		"FROM `person` "+
		"INNER JOIN `address` AS `address` ON `address`.`id` = `person`.`address_id` "+
		"WHERE `address`.`country` != ?", actual)
}

func TestDeleteQueryOuterJoin(t *testing.T) {
	assert := assert1.New(t)
	query := Delete(Person, Address).From(Person)
	query.OuterJoin(Left, Address).On(Address.ID, Equal, Person.AddressID)
	query.Where(Address.Country.NotEqual("US"))
	actual, values, err := query.SQL()
	assert.NoError(err)
	if assert.Equal(1, len(values)) {
		assert.Equal("US", values[0])
	}
	assert.Equal("DELETE `person`, `address` "+
		"FROM `person` "+
		"LEFT OUTER JOIN `address` AS `address` ON `address`.`id` = `person`.`address_id` "+
		"WHERE `address`.`country` != ?", actual)
}

func TestDeleteNoTablesSpecified(t *testing.T) {
	assert := assert1.New(t)
	query := Delete()
	query.Where(Person.ID.Equal(0))
	_, _, err := query.SQL()
	assert.EqualError(err, "at least one table must be specified to delete from")
}

func TestDeleteNoWhereClause(t *testing.T) {
	assert := assert1.New(t)
	query := Delete().From(Person)
	_, _, err := query.SQL()
	assert.EqualError(err, "delete requires a where clause")
}

func TestJoinWithNoCondition(t *testing.T) {
	assert := assert1.New(t)
	query := Delete().From(Person)
	query.InnerJoin(Address)
	query.Where(Person.AddressID.Equal(Address.ID))
	_, _, err := query.SQL()
	assert.EqualError(err, "no condition specified for join")
}
