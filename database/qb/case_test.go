package qb

import (
	"testing"

	assert1 "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CaseExpression(t *testing.T) {
	assert := assert1.New(t)
	require := require.New(t)
	expression := Case(Person.AddressID.Equal(Person.Name), 1).Else(0)
	expression.When(Person.AddressID.IsNull(), 2)
	actual, values := expression.ParameterizedSQL()
	assert.Equal("CASE WHEN `person`.`address_id` = `person`.`name` THEN ? WHEN `person`.`address_id` IS NULL THEN ? ELSE ? END", actual)
	require.Len(values, 3)
	assert.Equal(1, values[0])
	assert.Equal(2, values[1])
	assert.Equal(0, values[2])
}

func Test_CaseExpression_InSelect(t *testing.T) {
	assert := assert1.New(t)
	require := require.New(t)
	expression := Case(Person.AddressID.Equal(Person.Name), 1).Else(0)
	expression.When(Person.AddressID.IsNull(), 2)

	selectExpression := Select(
		Person.ID, expression.As("alias")).From(Person)
	actual, values, err := selectExpression.SQL(NewLimitOffset[int]().
		SetLimit(0).SetOffset(0))
	assert.Nil(err)
	assert.Equal("SELECT `person`.`id`, CASE WHEN `person`.`address_id` = `person`.`name` THEN ? WHEN `person`.`address_id` IS NULL THEN ? ELSE ? END AS `alias` FROM `person` AS `person`", actual)
	require.Len(values, 3)
	assert.Equal(1, values[0])
	assert.Equal(2, values[1])
	assert.Equal(0, values[2])
}
