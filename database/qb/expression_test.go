package qb

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
)

func TestExpresssionFieldToField(t *testing.T) {
	assert := assert1.New(t)
	expression := FieldComparison(Person.AddressID, Equal, Address.ID)
	sql, values := expression.SQL()
	assert.Empty(values)
	assert.Equal("`person`.`address_id` = `address`.`id`", sql)
}

func TestExpressionFieldToConstant(t *testing.T) {
	assert := assert1.New(t)
	expression := FieldComparison(Person.AddressID, Equal, SQLNow)
	sql, values := expression.SQL()
	assert.Empty(values)
	assert.Equal("`person`.`address_id` = NOW()", sql)
}

func TestExpresssionFieldComparison(t *testing.T) {
	assert := assert1.New(t)
	addressID := generator.TestID()
	expression := FieldComparison(Person.AddressID, Equal, addressID)
	sql, values := expression.SQL()
	if assert.Equal(1, len(values)) {
		assert.Equal(addressID, values[0])
	}

	assert.Equal("`person`.`address_id` = ?", sql)
}

func TestExpressionAnd(t *testing.T) {
	assert := assert1.New(t)
	expression := FieldComparison(Person.AddressID, Equal, Address.ID).And(FieldComparison(Address.Line, IsNot, nil))
	actual, values := expression.SQL()
	assert.Empty(values)
	assert.Equal("(`person`.`address_id` = `address`.`id` AND `address`.`line` IS NOT NULL)", actual)

	expression2 := FieldComparison(Person.AddressID, Equal, Address.ID).And(FieldComparison(Address.Line, Equal, Person.ID))
	actual, values = expression2.SQL()
	assert.Empty(values)
	assert.Equal("(`person`.`address_id` = `address`.`id` AND `address`.`line` = `person`.`id`)", actual)

	expression3 := expression.And(expression2)
	actual, values = expression3.SQL()
	assert.Empty(values)
	assert.Equal("((`person`.`address_id` = `address`.`id` AND `address`.`line` IS NOT NULL) AND (`person`.`address_id` = `address`.`id` AND `address`.`line` = `person`.`id`))", actual)
}

func TestExpressionOr(t *testing.T) {
	assert := assert1.New(t)
	expression := FieldComparison(Person.AddressID, Equal, Address.ID).Or(FieldComparison(Address.Line, IsNot, nil))
	actual, values := expression.SQL()
	assert.Empty(values)
	assert.Equal("(`person`.`address_id` = `address`.`id` OR `address`.`line` IS NOT NULL)", actual)

	expression2 := FieldComparison(Person.AddressID, Equal, Address.ID).Or(FieldComparison(Address.Line, Equal, Person.ID))
	actual, values = expression2.SQL()
	assert.Empty(values)
	assert.Equal("(`person`.`address_id` = `address`.`id` OR `address`.`line` = `person`.`id`)", actual)

	expression3 := expression.Or(expression2)
	actual, values = expression3.SQL()
	assert.Empty(values)
	assert.Equal("((`person`.`address_id` = `address`.`id` OR `address`.`line` IS NOT NULL) OR (`person`.`address_id` = `address`.`id` OR `address`.`line` = `person`.`id`))", actual)
}

func TestExpressionXor(t *testing.T) {
	assert := assert1.New(t)
	expression := FieldComparison(Person.AddressID, Equal, Address.ID).XOr(FieldComparison(Address.Line, IsNot, nil))
	actual, values := expression.SQL()
	assert.Empty(values)
	assert.Equal("(`person`.`address_id` = `address`.`id` XOR `address`.`line` IS NOT NULL)", actual)

	expression2 := FieldComparison(Person.AddressID, Equal, Address.ID).XOr(FieldComparison(Address.Line, Equal, Person.ID))
	actual, values = expression2.SQL()
	assert.Equal(0, len(values))
	assert.Equal("(`person`.`address_id` = `address`.`id` XOR `address`.`line` = `person`.`id`)", actual)

	expression3 := expression.XOr(expression2)
	actual, values = expression3.SQL()
	assert.Empty(values)
	assert.Equal("((`person`.`address_id` = `address`.`id` XOR `address`.`line` IS NOT NULL) XOR (`person`.`address_id` = `address`.`id` XOR `address`.`line` = `person`.`id`))", actual)
}

func TestExpressionMixed(t *testing.T) {
	assert := assert1.New(t)
	expression := FieldComparison(Person.AddressID, Equal, Address.ID).And(FieldComparison(Address.Line, IsNot, nil))
	actual, values := expression.SQL()
	assert.Empty(values)
	assert.Equal("(`person`.`address_id` = `address`.`id` AND `address`.`line` IS NOT NULL)", actual)
	address := generator.String(20)
	expression2 := FieldComparison(Person.AddressID, Equal, Address.ID).Or(FieldComparison(Address.Line, Equal, address))
	actual, values = expression2.SQL()
	if assert.Equal(1, len(values)) {
		assert.Equal(address, values[0])
	}
	assert.Equal("(`person`.`address_id` = `address`.`id` OR `address`.`line` = ?)", actual)

	expression3 := expression.XOr(expression2)
	actual, values = expression3.SQL()
	if assert.Equal(1, len(values)) {
		assert.Equal(address, values[0])
	}
	assert.Equal("((`person`.`address_id` = `address`.`id` AND `address`.`line` IS NOT NULL) XOR (`person`.`address_id` = `address`.`id` OR `address`.`line` = ?))", actual)
}

func TestExpressionMulti(t *testing.T) {
	assert := assert1.New(t)
	expression := FieldIn(Person.AddressID, "*", Address.ID, "foo")
	actual, values := expression.SQL()
	if assert.Equal(2, len(values)) {
		assert.Equal("*", values[0])
		assert.Equal("foo", values[1])
	}
	assert.Equal("`person`.`address_id` IN (?, `address`.`id`, ?)", actual)
}

func TestExpressionGetTables(t *testing.T) {
	assert := assert1.New(t)
	expression := FieldIn(Person.AddressID, "*", Address.ID, "foo")
	expression.And(FieldComparison(Person.AddressID, Equal, Address.ID))
	actual := expression.Tables()
	sql, _ := expression.SQL()
	assert.Equal("(`person`.`address_id` IN (?, `address`.`id`, ?) AND `person`.`address_id` = `address`.`id`)", sql)
	assert.Equal(4, len(actual))
	assert.Equal([]string{"person", "address", "person", "address"}, actual)
}
