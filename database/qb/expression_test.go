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

/*
the query we want to support building:

	select
	smtp_statistics.id,
	smtp_statistics.communication,
	count(smtp_message.id) as "sends",
	sum(case when smtp_message.status="EC_BOUNCE" then 1 else 0 end) as "bounces",
	sum(case when smtp_message.status="EC_COMPLAINT" then 1 else 0 end) as "complaints",
	sum(smtp_message.opens) as "opens",
	sum(case when (smtp_message.opens > 0) then 1 else 0 end) as "unique-opens"

	from smtp_statistics
	inner join smtp_message on smtp_statistics.communication = smtp_message.communication
	where smtp_statistics.id="smtp_stats_id";
*/
func Test_CaseSelect(t *testing.T) {
	assert := assert1.New(t)
	// create the case statement
	bounce := Case(SMTPMessage.Status.Equal("EC_BOUNCE"), 1).Else(0)
	complaint := Case(SMTPMessage.Status.Equal("EC_COMPLAINT"), 1).Else(0)
	uniqueOpens := Case(SMTPMessage.Opens.GreaterThan(0), 1).Else(0)
	// create the select statement
	selectStatement := Select(
		SMTPStatistics.ID,
		SMTPStatistics.Communication,
		Count(SMTPMessage.ID, "sends"),
		Sum(bounce, "bounces"),
		Sum(complaint, "complaints"),
		Sum(SMTPMessage.Opens, "opens"),
		Sum(uniqueOpens, "unique-opens")).
		From(SMTPStatistics)

	selectStatement.InnerJoin(SMTPMessage).On(SMTPStatistics.Communication, Equal, SMTPMessage.Communication)

	selectStatement.Where(FieldComparison(SMTPStatistics.ID, Equal, "smtp_stats_id"))

	actual, values, err := selectStatement.
		SQL(NewLimitOffset[int]().SetLimit(0).SetOffset(0))
	assert.Nil(err)
	assert.Equal("SELECT `smtp_statistics`.`id`, `smtp_statistics`.`communication`, COUNT(`smtp_message`.`id`) AS `sends`,"+
		" SUM(CASE WHEN `smtp_message`.`status` = ? THEN ? ELSE ? END) AS `bounces`, SUM(CASE WHEN `smtp_message`.`status` = ? "+
		"THEN ? ELSE ? END) AS `complaints`, SUM(`smtp_message`.`opens`) AS `opens`, SUM(CASE WHEN `smtp_message`.`opens` > ? THEN ? ELSE ? END) "+
		"AS `unique-opens` FROM `smtp_statistics` AS `smtp_statistics` INNER JOIN `smtp_message` AS `smtp_message` "+
		"ON `smtp_statistics`.`communication` = `smtp_message`.`communication` WHERE `smtp_statistics`.`id` = ? LIMIT 0", actual)
	assert.Equal(10, len(values))
	assert.Equal("EC_BOUNCE", values[0])
	assert.Equal(1, values[1])
	assert.Equal(0, values[2])
	assert.Equal("EC_COMPLAINT", values[3])
	assert.Equal(1, values[4])
	assert.Equal(0, values[5])
	assert.Equal(0, values[6])
	assert.Equal(1, values[7])
	assert.Equal(0, values[8])
	assert.Equal("smtp_stats_id", values[9])
}

type smtp_statistics struct {
	alias         string
	ID            TableField
	Communication TableField
	allColumns    TableField
}

func (s *smtp_statistics) GetName() string {
	return "smtp_statistics"
}

func (s *smtp_statistics) GetAlias() string {
	return s.alias
}

func (s *smtp_statistics) PrimaryKey() TableField {
	return s.ID
}

func (s *smtp_statistics) SortBy() (TableField, OrderDirection) {
	return s.ID, Ascending
}

func (s *smtp_statistics) AllColumns() TableField {
	return s.allColumns
}

func (s *smtp_statistics) ReadColumns() []TableField {
	return []TableField{
		s.ID,
		s.Communication,
	}
}

func (s *smtp_statistics) WriteColumns() []TableField {
	return s.ReadColumns()
}

func (s *smtp_statistics) Alias(alias string) *smtp_statistics {
	return &smtp_statistics{
		alias:         alias,
		ID:            TableField{Name: "id", Table: alias},
		Communication: TableField{Name: "communication", Table: alias},
		allColumns:    TableField{Name: "*", Table: alias},
	}
}

var SMTPStatistics = (&smtp_statistics{}).Alias("smtp_statistics")

type smtp_message struct {
	alias         string
	ID            TableField
	Communication TableField
	Status        TableField
	Opens         TableField
	allColumns    TableField
}

func (s *smtp_message) GetName() string {
	return "smtp_message"
}

func (s *smtp_message) GetAlias() string {
	return s.alias
}

func (s *smtp_message) PrimaryKey() TableField {
	return s.ID
}

func (s *smtp_message) SortBy() (TableField, OrderDirection) {
	return s.ID, Ascending
}

func (s *smtp_message) AllColumns() TableField {
	return s.allColumns
}

func (s *smtp_message) ReadColumns() []TableField {
	return []TableField{
		s.ID,
		s.Communication,
		s.Status,
		s.Opens,
	}
}

func (s *smtp_message) WriteColumns() []TableField {
	return s.ReadColumns()
}

func (s *smtp_message) Alias(alias string) *smtp_message {
	return &smtp_message{
		alias:         alias,
		ID:            TableField{Name: "id", Table: alias},
		Communication: TableField{Name: "communication", Table: alias},
		Status:        TableField{Name: "status", Table: alias},
		Opens:         TableField{Name: "opens", Table: alias},
		allColumns:    TableField{Name: "*", Table: alias},
	}
}

var SMTPMessage = (&smtp_message{}).Alias("smtp_message")
