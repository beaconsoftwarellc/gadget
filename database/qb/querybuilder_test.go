package qb

import (
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	assert1 "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type person struct {
	alias      string
	ID         TableField
	Name       TableField
	AddressID  TableField
	Age        TableField
	allColumns TableField
}

func (p *person) GetName() string {
	return "person"
}

func (p *person) GetAlias() string {
	return p.alias
}

func (p *person) PrimaryKey() TableField {
	return p.ID
}

func (p *person) SortBy() (TableField, OrderDirection) {
	return p.ID, Ascending
}

func (p *person) AllColumns() TableField {
	return p.allColumns
}

func (p *person) ReadColumns() []TableField {
	return []TableField{
		p.ID,
		p.Name,
		p.AddressID,
		p.Age,
	}
}

func (p *person) WriteColumns() []TableField {
	return p.ReadColumns()
}

func (p *person) Alias(alias string) *person {
	return &person{
		alias:      alias,
		ID:         TableField{Name: "id", Table: alias},
		Name:       TableField{Name: "name", Table: alias},
		AddressID:  TableField{Name: "address_id", Table: alias},
		Age:        TableField{Name: "age", Table: alias},
		allColumns: TableField{Name: "*", Table: alias},
	}
}

var Person = (&person{}).Alias("person")

type address struct {
	alias      string
	ID         TableField
	Line       TableField
	Line2      TableField
	Province   TableField
	Country    TableField
	allColumns TableField
}

func (a *address) GetName() string {
	return "address"
}

func (a *address) GetAlias() string {
	return a.alias
}

func (a *address) PrimaryKey() TableField {
	return a.ID
}

func (a *address) AllColumns() TableField {
	return a.allColumns
}

func (a *address) SortBy() (TableField, OrderDirection) {
	return a.ID, Ascending
}

func (a *address) ReadColumns() []TableField {
	return []TableField{
		a.ID,
		a.Line,
		a.Line2,
		a.Province,
		a.Country,
	}
}

func (a *address) WriteColumns() []TableField {
	return a.ReadColumns()
}

func (a *address) Alias(alias string) *address {
	return &address{
		alias:      alias,
		ID:         TableField{Name: "id", Table: alias},
		Line:       TableField{Name: "line", Table: alias},
		Line2:      TableField{Name: "line2", Table: alias},
		Province:   TableField{Name: "province", Table: alias},
		Country:    TableField{Name: "country", Table: alias},
		allColumns: TableField{Name: "*", Table: alias},
	}
}

var Address = (&address{}).Alias("address")

func TestQueryBuilderSimple(t *testing.T) {
	assert := assert1.New(t)
	query := Select(Person.ID, Person.Name).From(Person)
	actual, values, err := query.SQL(NoLimit, 0)
	assert.NoError(err)
	assert.Empty(values)
	expected := "SELECT `person`.`id`, `person`.`name` FROM `person` AS `person`"
	assert.Equal(expected, actual)
}

func TestQueryBuilderTableAlias(t *testing.T) {
	assert := assert1.New(t)
	table := Person.Alias("p")
	query := Select(table.ID, table.Name).From(table)
	actual, values, err := query.SQL(NoLimit, 0)
	assert.NoError(err)
	assert.Empty(values)
	expected := "SELECT `p`.`id`, `p`.`name` FROM `person` AS `p`"
	assert.Equal(expected, actual)
}

func TestQueryBuilderTableCount(t *testing.T) {
	assert := assert1.New(t)
	query := Select(Person.ID, Count(Person.Name, "person_name")).From(Person)
	actual, values, err := query.SQL(NoLimit, 0)
	assert.NoError(err)
	assert.Empty(values)
	expected := "SELECT `person`.`id`, COUNT(`person`.`name`) AS `person_name` FROM `person` AS `person`"
	assert.Equal(expected, actual)
}

func TestQueryBuilderTableSum(t *testing.T) {
	assert := assert1.New(t)
	query := Select(Person.ID, Sum(Person.Age, "person_age")).From(Person)
	actual, values, err := query.SQL(NoLimit, 0)
	assert.NoError(err)
	assert.Empty(values)
	expected := "SELECT `person`.`id`, SUM(`person`.`age`) AS `person_age` FROM `person` AS `person`"
	assert.Equal(expected, actual)
}

func TestQueryBuilderDistinct(t *testing.T) {
	assert := assert1.New(t)
	query := SelectDistinct(Person.ID, Person.Name).From(Person)
	actual, values, err := query.SQL(NoLimit, 0)
	assert.NoError(err)
	assert.Empty(values)
	expected := "SELECT DISTINCT `person`.`id`, `person`.`name` FROM `person` AS `person`"
	assert.Equal(expected, actual)
}

func TestQueryBuilderSelectQuery_Where(t *testing.T) {
	assert := assert1.New(t)
	query := Select(Person.ID, Person.Name).From(Person)
	query.Where(Person.ID.Equal(Person.Name))

	actual, values, err := query.SQL(NoLimit, 0)
	assert.Empty(values)
	assert.NoError(err)
	assert.Equal("SELECT `person`.`id`, `person`.`name` FROM `person` AS `person` WHERE `person`.`id` = `person`.`name`", actual)
}

func TestQueryBuilderSelectQuery_WhereAllConditions(t *testing.T) {
	assert := assert1.New(t)
	query := Select(Person.ID, Person.Name).From(Person)
	query.Where(Person.ID.Equal(Person.Name).
		And(Person.ID.NotEqual(Person.AddressID)).
		And(Person.ID.LessThan(1)).
		And(Person.ID.LessThanEqual(2)).
		And(Person.ID.GreaterThan(3)).
		And(Person.ID.GreaterThanEqual(4)).
		And(Person.ID.NullSafeEqual(Person.Name)).
		And(Person.ID.In(Person.Name)).
		And(Person.ID.Like(Person.Name)).
		And(Person.AddressID.IsNull()).
		And(Person.ID.IsNotNull()))

	actual, values, err := query.SQL(NoLimit, 0)
	assert.NotEmpty(values)
	assert.NoError(err)
	assert.Equal("SELECT `person`.`id`, `person`.`name` FROM `person` AS `person` "+
		"WHERE ((((((((((`person`.`id` = `person`.`name` "+
		"AND `person`.`id` != `person`.`address_id`) "+
		"AND `person`.`id` < ?) "+
		"AND `person`.`id` <= ?) "+
		"AND `person`.`id` > ?) "+
		"AND `person`.`id` >= ?) "+
		"AND `person`.`id` <=> `person`.`name`) "+
		"AND `person`.`id` = `person`.`name`) "+
		"AND `person`.`id` LIKE `person`.`name`) "+
		"AND `person`.`address_id` IS NULL) "+
		"AND `person`.`id` IS NOT NULL)", actual)
}

func TestQueryBuilderSelectQuery_WhereError(t *testing.T) {
	assert := assert1.New(t)
	query := Select(Person.ID, Person.Name).From(Person)
	query.Where(Person.AddressID.Equal(Address.ID))

	_, _, err := query.SQL(NoLimit, 0)
	assert.EqualError(err, NewMissingTablesError([]string{Address.GetName()}).Error())
}

func TestQueryBuilderSelectQuery_WhereValue(t *testing.T) {
	assert := assert1.New(t)
	where := Person.ID.Equal(12)
	query := Select(Person.ID, Person.Name).From(Person).Where(where)

	actual, values, err := query.SQL(NoLimit, 0)
	if assert.Equal(1, len(values)) {
		assert.Equal(12, values[0])
	}
	assert.NoError(err)
	assert.Equal("SELECT `person`.`id`, `person`.`name` FROM `person` AS `person` WHERE `person`.`id` = ?", actual)

	where.And(Person.Name.NotEqual("Jim Bob"))
	actual, values, err = query.SQL(12, 5)
	if assert.Equal(2, len(values)) {
		assert.Equal(12, values[0])
		assert.Equal("Jim Bob", values[1])
	}
	assert.NoError(err)
	assert.Equal("SELECT `person`.`id`, `person`.`name` "+
		"FROM `person` AS `person` "+
		"WHERE (`person`.`id` = ? AND `person`.`name` != ?) "+
		"LIMIT 12 OFFSET 5", actual)
}

func TestQueryBuilderJoin(t *testing.T) {
	assert := assert1.New(t)

	query := Select(Person.ID, Person.Name, Address.Line, Address.Country).From(Person)
	query.InnerJoin(Address).On(Person.AddressID, Equal, Address.ID)
	query.Where(Person.Name.NotEqual("Jim").And(FieldComparison(Address.ID, NotEqual, 12)))

	actual, values, err := query.SQL(10, 0)
	if assert.Equal(2, len(values)) {
		assert.Equal("Jim", values[0])
		assert.Equal(12, values[1])
	}
	assert.NoError(err)
	assert.Equal("SELECT `person`.`id`, `person`.`name`, `address`.`line`, `address`.`country` "+
		"FROM `person` AS `person` "+
		"INNER JOIN `address` AS `address` ON `person`.`address_id` = `address`.`id` "+
		"WHERE (`person`.`name` != ? AND `address`.`id` != ?) "+
		"LIMIT 10 OFFSET 0", actual)
}

func TestQueryBuilderJoin_SQL_Outer(t *testing.T) {
	assert := assert1.New(t)

	query := Select(Person.ID, Person.Name, Address.Line, Address.Country).From(Person)
	query.OuterJoin(Left, Address).On(Person.AddressID, Equal, Address.ID)
	query.Where(Person.Name.NotEqual("Jim").And(FieldComparison(Address.ID, NotEqual, 12)))

	actual, values, err := query.SQL(10, 0)
	if assert.Equal(2, len(values)) {
		assert.Equal("Jim", values[0])
		assert.Equal(12, values[1])
	}
	assert.NoError(err)
	assert.Equal("SELECT `person`.`id`, `person`.`name`, `address`.`line`, `address`.`country` "+
		"FROM `person` AS `person` "+
		"LEFT OUTER JOIN `address` AS `address` ON `person`.`address_id` = `address`.`id` "+
		"WHERE (`person`.`name` != ? AND `address`.`id` != ?) "+
		"LIMIT 10 OFFSET 0", actual)
}

func TestQueryBuilderJoin_OnValue(t *testing.T) {
	assert := assert1.New(t)

	query := Select(Person.ID, Person.Name, Address.Line, Address.Country).From(Person)
	query.OuterJoin(Left, Address).On(Address.ID, Equal, "Bob")
	query.Where(Person.Name.NotEqual("Jim").And(FieldComparison(Address.ID, NotEqual, 12)))

	actual, values, err := query.SQL(10, 0)
	assert.NoError(err)
	if assert.Equal(3, len(values)) {
		assert.Equal("Bob", values[0])
		assert.Equal("Jim", values[1])
		assert.Equal(12, values[2])
	}
	assert.Equal("SELECT `person`.`id`, `person`.`name`, `address`.`line`, `address`.`country` "+
		"FROM `person` AS `person` "+
		"LEFT OUTER JOIN `address` AS `address` ON `address`.`id` = ? "+
		"WHERE (`person`.`name` != ? AND `address`.`id` != ?) "+
		"LIMIT 10 OFFSET 0", actual)
}

func TestQueryBuilderSelectQuery_SQL_JoinFVError(t *testing.T) {
	assert := assert1.New(t)

	query := Select(Person.ID, Person.Name, Address.Line, Address.Country).From(Person)
	query.OuterJoin(Left, Address).On(Person.AddressID, Equal, "Bob")
	query.Where(Person.Name.NotEqual("Jim").And(FieldComparison(Address.ID, NotEqual, 12)))
	_, _, err := query.SQL(NoLimit, 0)
	assert.EqualError(err, (&JoinError{joinTable: Address.GetName(), conditionTables: []string{Person.GetName()}}).Error())
}

func TestQueryBuilderSelectQuery_SQL_JoinFFError(t *testing.T) {
	assert := assert1.New(t)

	query := Select(Person.ID, Person.Name, Address.Line, Address.Country).From(Person)
	query.OuterJoin(Left, Address).On(Person.AddressID, Equal, Person.ID)
	query.Where(Person.Name.NotEqual("Jim").And(FieldComparison(Address.ID, NotEqual, 12)))
	_, _, err := query.SQL(NoLimit, 0)
	assert.EqualError(err, (&JoinError{joinTable: Address.GetName(), conditionTables: []string{Person.GetName(), Person.GetName()}}).Error())
}

func TestQueryBuilderOrderBy_SQL(t *testing.T) {
	assert := assert1.New(t)

	actual, values, err := Select(Person.ID, Person.Name).From(Person).OrderBy(Person.ID, Ascending).SQL(10, 10)
	assert.NoError(err)
	assert.Empty(values)
	assert.Equal("SELECT `person`.`id`, `person`.`name` FROM `person` AS `person` ORDER BY `person`.`id` ASC LIMIT 10 OFFSET 10", actual)
}

func TestQueryBuilderOrderByMulti_SQL(t *testing.T) {
	assert := assert1.New(t)
	query := Select(Person.ID, Person.Name).From(Person).OrderBy(Person.ID, Ascending)
	query.OrderBy(Person.Name, Descending)
	actual, values, err := query.SQL(10, 10)
	assert.NoError(err)
	assert.Empty(values)
	assert.Equal("SELECT `person`.`id`, `person`.`name` FROM `person` AS `person` ORDER BY `person`.`id` ASC, `person`.`name` DESC LIMIT 10 OFFSET 10", actual)
}

func TestQueryBuilderFromNotSetError(t *testing.T) {
	assert := assert1.New(t)
	query := Select(Person.ID)
	query.Where(Person.ID.Equal(3))
	_, _, err := query.SQL(NoLimit, 0)
	assert.EqualError(err, NewValidationFromNotSetError().Error())
}

func TestQueryBuilderAlias(t *testing.T) {
	assert := assert1.New(t)
	query := Select(Person.ID, Alias(Person.Name, "person_name")).From(Person)
	actual, values, err := query.SQL(NoLimit, 10)
	assert.NoError(err)
	assert.Empty(values)
	assert.Equal("SELECT `person`.`id`, `person`.`name` AS `person_name` FROM `person` AS `person`", actual)
}

func TestQueryBuilderCoalesce(t *testing.T) {
	assert := assert1.New(t)
	value := generator.ID("value")
	query := Select(Person.ID, Coalesce(Person.Name, value, "coalesced")).From(Person)
	actual, values, err := query.SQL(NoLimit, 10)
	assert.NoError(err)
	assert.Len(values, 1)
	assert.Equal(value, values[0])
	assert.Equal("SELECT `person`.`id`, COALESCE(`person`.`name`, ?) AS `coalesced` FROM `person` AS `person`", actual)
}

func TestQueryBuilderIfFieldCondition(t *testing.T) {
	assert := assert1.New(t)
	require := require.New(t)
	query := Select(Person.ID, If(Person.Name.Equal(Person.ID), "1", "0", "has_robot_name")).From(Person)
	actual, values, err := query.SQL(NoLimit, 10)
	assert.NoError(err)
	require.Len(values, 2)
	assert.Equal("1", values[0])
	assert.Equal("0", values[1])
	assert.Equal("SELECT `person`.`id`, IF(`person`.`name` = `person`.`id`, ?, ?) AS `has_robot_name` FROM `person` AS `person`", actual)
}

func TestQueryBuilderIfStringCondition(t *testing.T) {
	assert := assert1.New(t)
	require := require.New(t)
	query := Select(Person.ID, If(Person.Name.Equal("Joe"), "1", "0", "is_joe")).From(Person)
	actual, values, err := query.SQL(NoLimit, 10)
	assert.NoError(err)
	require.Len(values, 3)
	assert.Equal("Joe", values[0])
	assert.Equal("1", values[1])
	assert.Equal("0", values[2])
	assert.Equal("SELECT `person`.`id`, IF(`person`.`name` = ?, ?, ?) AS `is_joe` FROM `person` AS `person`", actual)
}

func TestQueryBuilderGroupBy(t *testing.T) {
	assert := assert1.New(t)
	query := Select(Person.ID, Person.Name, Person.AddressID).From(Person).GroupBy(Person.Name, Person.AddressID)
	actual, values, err := query.SQL(NoLimit, 10)
	assert.NoError(err)
	assert.Empty(values)
	assert.Equal("SELECT `person`.`id`, `person`.`name`, `person`.`address_id` FROM `person` AS `person`"+
		" GROUP BY `person`.`name`, `person`.`address_id`", actual)
}

func TestSelectNotNull(t *testing.T) {
	assert := assert1.New(t)
	query := Select(NotNull(Person.ID, "person_id_not_null"), Person.Name)
	query.From(Person)
	actual, values, err := query.SQL(NoLimit, 10)
	assert.NoError(err)
	assert.Empty(values)
	assert.Equal("SELECT (`person`.`id` IS NOT NULL) AS `person_id_not_null`, `person`.`name` FROM `person` AS `person`", actual)
}

func TestSelectBitwise(t *testing.T) {
	assert := assert1.New(t)
	query := Select(Person.Name)
	query.From(Person)
	query.Where(Person.ID.NotEqual(Bitwise(Person.ID, BitwiseAnd, 5)))
	actual, values, err := query.SQL(NoLimit, 10)
	assert.NoError(err)
	assert.Equal(values, []any{5})
	assert.Equal("SELECT `person`.`name` FROM `person` AS `person` WHERE `person`.`id` != `person`.`id` & ?", actual)
}

func TestSelectFrom(t *testing.T) {
	assert := assert1.New(t)
	query := Select(Person.ID, Person.Name, Person.AddressID).From(Person).GroupBy(Person.Name, Person.AddressID)
	query2 := query.SelectFrom(Person.ID)

	actual, values, err := query.SQL(NoLimit, 10)
	assert.NoError(err)
	assert.Empty(values)
	assert.Equal("SELECT `person`.`id`, `person`.`name`, `person`.`address_id` FROM `person` AS `person`"+
		" GROUP BY `person`.`name`, `person`.`address_id`", actual)

	actual, values, err = query2.SQL(NoLimit, 10)
	assert.NoError(err)
	assert.Empty(values)
	assert.Equal("SELECT `person`.`id` FROM `person` AS `person`"+
		" GROUP BY `person`.`name`, `person`.`address_id`", actual)
}

func TestUpdateBitwise(t *testing.T) {
	assert := assert1.New(t)

	actual, values, err := Update(Person).Set(Person.AddressID,
		Bitwise(Person.AddressID, BitwiseAndNegation, "5")).SQL(NoLimit)
	assert.NoError(err)
	assert.Equal(values, []any{"5"})
	assert.Equal("UPDATE `person` SET  `person`.`address_id` = `person`.`address_id` &~ ?", actual)
}
