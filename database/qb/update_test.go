package qb

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/generator"
)

func TestUpdateQuery(t *testing.T) {
	assert := assert1.New(t)
	expectedName := generator.String(20)
	query := Update(Person).Set(Person.Name, expectedName)
	sql, values, err := query.SQL(0)
	assert.NoError(err)
	if assert.Equal(1, len(values)) {
		assert.Equal(expectedName, values[0])
	}
	assert.Equal("UPDATE `person` SET  `person`.`name` = ?", sql)
}

func TestUpdateQueryParameterized(t *testing.T) {
	assert := assert1.New(t)
	query := Update(Person).SetParam(Person.Name).Where(Person.ID.Equal(":" + Person.ID.GetName()))
	sql, err := query.ParameterizedSQL(0)
	assert.NoError(err)
	assert.Equal("UPDATE `person` SET  `person`.`name` = :name WHERE `person`.`id` = :id", sql)
}

func TestUpdateQueryMulitpleFields(t *testing.T) {
	assert := assert1.New(t)
	expectedName := generator.String(20)
	expectedAddressID := generator.TestID()
	query := Update(Person).Set(Person.Name, expectedName).Set(Person.AddressID, expectedAddressID)
	sql, values, err := query.SQL(0)
	assert.NoError(err)
	if assert.Equal(2, len(values)) {
		assert.Equal(expectedName, values[0])
		assert.Equal(expectedAddressID, values[1])
	}
	assert.Equal("UPDATE `person` SET  `person`.`name` = ?, `person`.`address_id` = ?", sql)
}

func TestUpdateQueryWhere(t *testing.T) {
	assert := assert1.New(t)
	expectedName := generator.String(20)
	expectedAddressID := generator.TestID()
	query := Update(Person).Set(Person.Name, expectedName).Where(Person.AddressID.Equal(expectedAddressID))
	sql, values, err := query.SQL(0)
	assert.NoError(err)
	if assert.Equal(2, len(values)) {
		assert.Equal(expectedName, values[0])
		assert.Equal(expectedAddressID, values[1])
	}
	assert.Equal("UPDATE `person` SET  `person`.`name` = ? WHERE `person`.`address_id` = ?", sql)
}

func TestUpdateQueryOrderBy(t *testing.T) {
	assert := assert1.New(t)
	expectedName := generator.String(20)
	expectedAddressID := generator.TestID()
	query := Update(Person).Set(Person.Name, expectedName)
	query.Where(Person.AddressID.Equal(expectedAddressID))
	query.OrderBy(Person.Name, Descending)
	sql, values, err := query.SQL(0)
	assert.NoError(err)
	if assert.Equal(2, len(values)) {
		assert.Equal(expectedName, values[0])
		assert.Equal(expectedAddressID, values[1])
	}
	assert.Equal("UPDATE `person` SET  `person`.`name` = ? WHERE `person`.`address_id` = ? ORDER BY `name` DESC", sql)
}

func TestUpdateQueryWhereOrderByLimit(t *testing.T) {
	assert := assert1.New(t)
	expectedName := generator.String(20)
	expectedAddressID := generator.TestID()
	query := Update(Person).Set(Person.Name, expectedName)
	query.Where(Person.AddressID.Equal(expectedAddressID))
	query.OrderBy(Person.Name, Descending)
	sql, values, err := query.SQL(10)
	assert.NoError(err)
	if assert.Equal(2, len(values)) {
		assert.Equal(expectedName, values[0])
		assert.Equal(expectedAddressID, values[1])
	}
	assert.Equal("UPDATE `person` SET  `person`.`name` = ? WHERE `person`.`address_id` = ? ORDER BY `name` DESC LIMIT 10", sql)
}
