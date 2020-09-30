package database

import (
	"fmt"
	"os"
	"path"

	"github.com/mattes/migrate"
	_ "github.com/mattes/migrate/database/mysql" // imported for side effect as driver for mysql
	_ "github.com/mattes/migrate/source/file"    // imported for side effect as driver for migrate

	"github.com/beaconsoftwarellc/gadget/fileutil"
	"github.com/beaconsoftwarellc/gadget/generator"
	"github.com/beaconsoftwarellc/gadget/log"
)

// generateSQLFiles writes temporary files from the migration map
func generateSQLFiles(migrations map[string]string) (string, error) {
	basepath := path.Join(os.TempDir(), "test_db", generator.ID("dbm"))
	_, err := fileutil.EnsureDir(basepath, 0777)

	if nil != err {
		return "", fmt.Errorf("Unable to create directory %s\n%s", basepath, err)
	}

	for filename, data := range migrations {
		f, err := os.Create(path.Join(basepath, filename))
		if nil != err {
			return "", fmt.Errorf("Unable to write %s to %s for database migrations\n%s", filename, basepath, err)
		}
		defer f.Close()
		f.WriteString(data)
	}

	return fmt.Sprintf("file://%s", basepath), nil
}

// Migrate ensures that the database is up to date
// Panics on error since this is an unrecoverable, fatal issue
func Migrate(migrations map[string]string, dbURL string) {
	sqlFilesPath, err := generateSQLFiles(migrations)
	if err != nil {
		panic(log.Fatal(err))
	}
	m, err := migrate.New(sqlFilesPath, dbURL)
	if err != nil {
		msg := fmt.Sprintf("couldn't load migration scripts from %s, %s (%s)", sqlFilesPath, dbURL, err)
		log.Fatalf(msg)
		panic(msg)
	}
	defer m.Close()

	// Migrate all the way up ...
	if err := m.Up(); err != nil && "no change" != err.Error() {
		msg := fmt.Sprintf("Migration ERROR: %#v", err)
		log.Fatalf(msg)
		panic(msg)
	}
}

// Reset runs all rollback migrations for the database
// Panics on error since this is an unrecoverable, fatal issue
// This will essentially nuke your database.  Only really useful for test scenario cleanup.
func Reset(migrations map[string]string, dbURL string) {
	sqlFilesPath, err := generateSQLFiles(migrations)
	if err != nil {
		panic(log.Error(err))
	}
	m, err := migrate.New(sqlFilesPath, dbURL)
	if err != nil {
		msg := fmt.Sprintf("couldn't load migration scripts from %s (%s)", sqlFilesPath, err)
		log.Fatalf(msg)
		panic(msg)
	}
	defer m.Close()

	// Migrate all the way down ...
	if err := m.Down(); err != nil {
		msg := fmt.Sprintf("Rollback ERROR: %#v", err)
		log.Fatalf(msg)
		panic(msg)
	}
}
