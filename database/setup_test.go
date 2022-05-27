package database

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/environment"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
)

type specification struct {
	DatabaseType string
	DatabaseURL  string `env:"TST_DATABASE_URL"`
	DB           *Database
}

func (spec specification) NumberOfRetries() int {
	return 5
}

func (spec specification) WaitBetweenRetries() time.Duration {
	return 1 * time.Millisecond
}

// DatabaseConnection return the connection string for the database
func (spec specification) DatabaseConnection() string {
	return spec.DatabaseURL
}

// DatabaseDialect returns the dialect (mysql, mongo, etc) for the connection
func (spec specification) DatabaseDialect() string {
	return spec.DatabaseType
}

// DatabaseDialectURL returns the dialect (mysql, mongo, etc) for the connection
func (spec specification) DatabaseDialectURL() string {
	return fmt.Sprintf("%s://%s", spec.DatabaseType, spec.DatabaseURL)
}

func newSpecification() *specification {
	config := &specification{
		DatabaseType: "mysql",
		DatabaseURL:  "test:test@/test_db?parseTime=true",
	}
	environment.Process(config)
	config.DB = Initialize(config)
	return config
}

// testMeta defines a table
type testMeta struct {
	alias string
	ID    qb.TableField
	Name  qb.TableField
	Place qb.TableField

	CreatedOn qb.TableField
	UpdatedOn qb.TableField
}

func (p *testMeta) GetName() string {
	return "test_record"
}

func (p *testMeta) GetAlias() string {
	return p.alias
}

func (p *testMeta) PrimaryKey() qb.TableField {
	return p.ID
}

func (p *testMeta) AllColumns() qb.TableField {
	return qb.TableField{Table: p.GetName(), Name: "*"}
}

func (p *testMeta) SortBy() (qb.TableField, qb.OrderDirection) {
	return p.CreatedOn, qb.Ascending
}

func (p *testMeta) ReadColumns() []qb.TableField {
	return []qb.TableField{
		p.ID,
		p.Name,
		p.Place,
		p.CreatedOn,
		p.UpdatedOn,
	}
}

func (p *testMeta) WriteColumns() []qb.TableField {
	return []qb.TableField{
		p.Name,
		p.Place,
	}
}

func (p *testMeta) Alias(alias string) *testMeta {
	return &testMeta{
		alias: alias,
		ID:    qb.TableField{Name: "id", Table: alias},
		Name:  qb.TableField{Name: "name", Table: alias},
		Place: qb.TableField{Name: "place", Table: alias},

		CreatedOn: qb.TableField{Name: "created_on", Table: alias},
		UpdatedOn: qb.TableField{Name: "updated_on", Table: alias},
	}
}

var TestMeta = (&testMeta{}).Alias("test_record")

type TestRecord struct {
	DefaultRecord
	ID        string         `db:"id,read_only"`
	Name      string         `db:"name"`
	Place     sql.NullString `db:"place"`
	CreatedOn time.Time      `db:"created_on,read_only"`
	UpdatedOn time.Time      `db:"updated_on,read_only"`
	Skip      string
}

func (record *TestRecord) Initialize() {
	record.ID = generator.ID("tst")
}

func (record *TestRecord) PrimaryKey() PrimaryKeyValue {
	return NewPrimaryKey(record.ID)
}

func (record *TestRecord) Meta() qb.Table {
	return TestMeta
}

// detailsTestMeta defines a table
type detailsTestMeta struct {
	alias string
	ID    qb.TableField
	Name  qb.TableField
}

func (p *detailsTestMeta) GetName() string {
	return "details_test_record"
}

func (p *detailsTestMeta) GetAlias() string {
	return p.alias
}

func (p *detailsTestMeta) PrimaryKey() qb.TableField {
	return p.ID
}

func (p *detailsTestMeta) AllColumns() qb.TableField {
	return qb.TableField{Table: p.GetName(), Name: "*"}
}

func (p *detailsTestMeta) SortBy() (qb.TableField, qb.OrderDirection) {
	return p.ID, qb.Ascending
}

func (p *detailsTestMeta) ReadColumns() []qb.TableField {
	return p.WriteColumns()
}

func (p *detailsTestMeta) WriteColumns() []qb.TableField {
	return []qb.TableField{
		p.ID,
		p.Name,
	}
}

func (p *detailsTestMeta) Alias(alias string) *detailsTestMeta {
	return &detailsTestMeta{
		alias: alias,
		ID:    qb.TableField{Name: "id", Table: alias},
		Name:  qb.TableField{Name: "name", Table: alias},
	}
}

var DetailsTestMeta = (&detailsTestMeta{}).Alias("details_test_record")

type DetailsTestRecord struct {
	ID   string `db:"id,nonsense"`
	Name string `db:"name"`
}

func (record DetailsTestRecord) Meta() qb.Table {
	return DetailsTestMeta
}

func (record DetailsTestRecord) Initialize() {
}

func (record DetailsTestRecord) Key() string {
	return "name"
}

func (record DetailsTestRecord) PrimaryKey() PrimaryKeyValue {
	return NewPrimaryKey(record.ID)
}

// testDuperMeta defines a table
type testDuperMeta struct {
	alias string
	ID    qb.TableField
}

func (p *testDuperMeta) GetName() string {
	return "test_duper"
}

func (p *testDuperMeta) GetAlias() string {
	return p.alias
}

func (p *testDuperMeta) PrimaryKey() qb.TableField {
	return p.ID
}

func (p *testDuperMeta) AllColumns() qb.TableField {
	return qb.TableField{Table: p.GetName(), Name: "*"}
}

func (p *testDuperMeta) SortBy() (qb.TableField, qb.OrderDirection) {
	return p.ID, qb.Ascending
}

func (p *testDuperMeta) ReadColumns() []qb.TableField {
	return p.WriteColumns()
}

func (p *testDuperMeta) WriteColumns() []qb.TableField {
	return []qb.TableField{p.ID}
}

func (p *testDuperMeta) Alias(alias string) *testDuperMeta {
	return &testDuperMeta{
		alias: alias,
		ID:    qb.TableField{Name: "id", Table: alias},
	}
}

var TestDuperMeta = (&testDuperMeta{}).Alias("test_duper")

type TestDuper struct {
	ID         string `db:"id"`
	intializer func() string
}

func (record *TestDuper) Meta() qb.Table {
	return TestDuperMeta
}

func (record *TestDuper) Initialize() {
	log.Debugf("\n\nInitializing: %s", record.ID)
	record.ID = record.intializer()
	log.Debugf("Initialized: %s\n\n", record.ID)

}

func (record *TestDuper) Key() string {
	return "id"
}

func (record *TestDuper) PrimaryKey() PrimaryKeyValue {
	return NewPrimaryKey(record.ID)
}

func initialize() string {
	return generator.ID("dup")
}

func NewTestDuper() *TestDuper {
	return &TestDuper{
		intializer: initialize,
	}
}

func TestMain(m *testing.M) {
	config := newSpecification()
	migrations := make(map[string]string)
	migrations["0001_.up.sql"] = `CREATE TABLE IF NOT EXISTS test_record (
			id varchar(128) primary key,
			name varchar(128) not null unique,
			place varchar(128) null,
			created_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS test_duper (
			id varchar(128) primary key
		);
`
	migrations["0001_foo.down.sql"] = `DROP TABLE IF EXISTS test_record;
	DROP TABLE IF EXISTS test_duper;
`
	Migrate(migrations, config.DatabaseDialectURL())

	res := m.Run()
	Reset(migrations, config.DatabaseDialectURL())

	os.Exit(res)
}
