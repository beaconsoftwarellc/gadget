package database

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/environment"
)

func TestGenerateSqlFiles(t *testing.T) {
	assert := assert1.New(t)
	migrations := make(map[string]string)

	basedir := path.Join(os.TempDir(), "test_db")

	os.RemoveAll(basedir)
	// Create the target directory as a file instead of a directory
	os.Create(basedir)
	actual, err := generateSQLFiles(migrations)
	// expect a error to the tune of 'cannot create directory'
	assert.Equal("", actual)
	assert.Error(err)
	// Triggers first panic since it cannot create the migration files
	assert.Panics(func() { Migrate(migrations, "") })
	assert.Panics(func() { Reset(migrations, "") })
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
	assert.Panics(func() { Migrate(migrations, "") })
	assert.Panics(func() { Reset(migrations, "") })

	config := &specification{
		DatabaseType: "mysql",
	}
	environment.Process(config)

	// Triggers second panic since it cannot find any files due to the name not conforming to 00_stuff.up.sql
	assert.Panics(func() { Migrate(migrations, config.DatabaseDialectURL()) })
	assert.Panics(func() { Reset(migrations, config.DatabaseDialectURL()) })
}
