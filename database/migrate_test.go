package database

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/v2/environment"
	"github.com/beaconsoftwarellc/gadget/v2/log"
)

func TestGenerateSqlFiles(t *testing.T) {
	assert := assert1.New(t)
	migrations := make(map[string]string)

	basedir := path.Join(os.TempDir(), "db_migrations")

	os.RemoveAll(basedir)
	// Create the target directory as a file instead of a directory
	os.Create(basedir)
	actual, err := generateSQLFiles(migrations)
	// expect a error to the tune of 'cannot create directory'
	assert.Equal("", actual)
	assert.Error(err)
	// Triggers first panic since it cannot create the migration files
	assert.Panics(func() { Migrate(migrations, "", log.NewStackLogger()) })
	assert.Panics(func() { Reset(migrations, "", log.NewStackLogger()) })
	os.RemoveAll(basedir)

	// Assert that a directory was created even if migrations is empty
	actual, err = generateSQLFiles(migrations)
	assert.NoError(err)
	assert.NotEmpty(actual)

	filename := "foo.txt"
	expected := "this is a test"
	migrations[filename] = expected
	actual, err = generateSQLFiles(migrations)
	assert.NoError(err)
	assert.NotEmpty(actual)
	content, err := ioutil.ReadFile(path.Join(strings.Replace(actual, "file://", "", 1), filename))
	assert.NoError(err)
	assert.Equal(expected, string(content))
}

func TestMigrateAndResetErrors(t *testing.T) {
	assert := assert1.New(t)
	migrations := make(map[string]string)
	migrations["0002.up.sql"] = `CREATE TABLE test_create (
			id varchar(128) primary key
		);
`
	migrations["0002.down.sql"] = `DROP TABLE IF EXISTS test_create;
`

	// Triggers second panic since it cannot connect to the database
	assert.Panics(func() { Migrate(migrations, "", log.NewStackLogger()) })
	assert.Panics(func() { Reset(migrations, "", log.NewStackLogger()) })

	config := &specification{
		DatabaseType: "mysql",
	}
	environment.Process(config, log.NewStackLogger())

	// Triggers second panic since it cannot find any files due to the name not conforming to 00_stuff.up.sql
	assert.Panics(func() { Migrate(migrations, config.DatabaseDialectURL(), log.NewStackLogger()) })
	assert.Panics(func() { Reset(migrations, config.DatabaseDialectURL(), log.NewStackLogger()) })
}
