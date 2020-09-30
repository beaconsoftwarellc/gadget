package database

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattes/migrate/database/mysql" // imported for side effect as driver for mysql
	_ "github.com/mattes/migrate/source/file"    // imported for side effect as driver for migrate
	yaml "gopkg.in/yaml.v3"

	"github.com/beaconsoftwarellc/gadget/database/qb"
	"github.com/beaconsoftwarellc/gadget/log"
	"github.com/beaconsoftwarellc/gadget/stringutil"
)

const (
	mysqlTimestampFormat = "2006-01-02 15:04:05.999999"
	inputTimestampFormat = "2006-01-02T15:04:05Z"
)

// Bootstrap runs sql commands to import baseline data into the database
func Bootstrap(db *Database, lines []string, logger log.Logger) {
	for _, sql := range lines {
		if _, err := db.Exec(sql); nil != err {
			logger.Errorf("Error bootstrapping %s\n%#v", sql, err)
		}
	}
}

// Bootstrapper handles reading/writing yaml and sql files for data bootstrapping
type Bootstrapper interface {
	// ReadYaml reads a yaml file and exit on error
	ReadYaml(filename string, target interface{})
	// WriteYaml writes yaml to a file and exit on error
	WriteYaml(filename string, output interface{})
	// WriteSQL writes the sql version of a yaml file
	WriteSQL(filename string, output []string)
	// WriteConstants writes the constants for the data
	WriteConstants(filename string, output map[string]map[string]string)
	// Insert a record into the database
	Insert(record Record) error
	// UpsertQuery generates SQL to insert / update a record
	UpsertQuery(record Record) string
	// DB returns a *Database instances
	DB() *Database
	// TX returns the database transaction
	TX() *sqlx.Tx
	// FailOnError will rollback transaction and exit if an error is received
	FailOnError(err error)
}

type bootstrapper struct {
	sqlOrder int
	db       *Database
	tx       *sqlx.Tx
	log      log.Logger
}

// NewBootstrapper returns the primary implementation of the Bootstrapper interface
func NewBootstrapper(db *Database) Bootstrapper {
	return &bootstrapper{
		db:  db,
		tx:  db.MustBegin(),
		log: log.New("bootstrap", log.FunctionFromEnv()),
	}
}

func (bs *bootstrapper) DB() *Database {
	return bs.db
}

func (bs *bootstrapper) TX() *sqlx.Tx {
	return bs.tx
}

func (bs *bootstrapper) FailOnError(err error) {
	if nil != err {
		bs.log.Error(err)
		bs.tx.Rollback()
		os.Exit(1)
	}
}

func (bs *bootstrapper) ReadYaml(filename string, target interface{}) {
	data, err := ioutil.ReadFile(filename)
	bs.FailOnError(err)
	err = yaml.Unmarshal(data, target)
	bs.FailOnError(err)
}

func (bs *bootstrapper) WriteYaml(filename string, output interface{}) {
	data, err := yaml.Marshal(output)
	bs.FailOnError(err)
	bs.FailOnError(ioutil.WriteFile(filename, data, 0700))
}

const constantsTemplate = `
package constants

// THIS IS A GENERATED FILE. DO NOT MODIFY

{{ range $type := .Categories }}
// Constants related to {{$type}}
const ({{ range $name := index $.Constants $type }}
	// {{$type}}{{$name}} represents the {{index $.Data $type $name}} {{$type}} from the database
	{{$type}}{{$name}} = "{{index $.Data $type $name}}"{{end}}
)
{{end}}
`

func (bs *bootstrapper) WriteConstants(filename string, output map[string]map[string]string) {
	data := newSortedData(output)
	t := template.Must(template.New("constants").Parse(constantsTemplate))
	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	bs.FailOnError(err)
	bs.FailOnError(t.Execute(fd, data))
}

type sortedData struct {
	Data       map[string]map[string]string
	Categories []string
	Constants  map[string][]string
}

func newSortedData(data map[string]map[string]string) *sortedData {
	d := &sortedData{
		Data:       data,
		Categories: make([]string, 0, len(data)),
		Constants:  map[string][]string{},
	}
	for k, constants := range data {
		d.Categories = append(d.Categories, k)
		d.Constants[k] = make([]string, 0, len(constants))
		for c := range constants {
			d.Constants[k] = append(d.Constants[k], c)
		}
		sort.Strings(d.Constants[k])
	}
	sort.Strings(d.Categories)
	return d
}

func (bs *bootstrapper) WriteSQL(filename string, output []string) {
	out := []byte(strings.Join(output, "\n"))
	filename = fmt.Sprintf("%d.%s", bs.sqlOrder, strings.Replace(filename, ".yaml", ".sql", 1))
	bs.sqlOrder++
	bs.FailOnError(ioutil.WriteFile(filename, out, 0700))
}

func (bs *bootstrapper) Insert(record Record) error {
	if "" != record.PrimaryKey().Value() {
		err := bs.db.UpsertTx(record, bs.TX())
		return err
	}
	return bs.db.CreateTx(record, bs.TX())
}

func (bs *bootstrapper) toSQLString(val interface{}) string {
	if v, ok := val.(float64); ok {
		return strconv.Itoa(int(v))
	}
	if v, ok := val.(bool); ok {
		return strconv.FormatBool(v)
	}
	if v, ok := val.(map[string]interface{}); ok {
		nullTime := &mysql.NullTime{}
		nullTime.Scan(v["Time"])
		if nullTime.Valid {
			return fmt.Sprintf("'%s'", nullTime.Time.Format(mysqlTimestampFormat))
		}
		return "null"
	}
	emptyTime := time.Time{}
	v, _ := time.Parse(inputTimestampFormat, val.(string))
	if v == emptyTime {
		return fmt.Sprintf("'%s'", strings.Replace(val.(string), "'", "\\'", -1))
	}
	return fmt.Sprintf("'%s'", v.Format(mysqlTimestampFormat))
}

func (bs *bootstrapper) UpsertQuery(record Record) string {
	columnValues := map[string]interface{}{}
	tmp, _ := json.Marshal(record)
	json.Unmarshal(tmp, &columnValues)
	for k, v := range columnValues {
		k = stringutil.Underscore(k)
		columnValues[k] = v
	}

	insertCols := appendIfMissing(record.Meta().ReadColumns(), record.Meta().PrimaryKey())
	insertVals := make([]interface{}, len(insertCols))
	for i, col := range insertCols {
		insertVals[i] = columnValues[col.GetName()]
	}
	updateCols := make([]qb.TableField, len(record.Meta().WriteColumns()))
	copy(updateCols, record.Meta().WriteColumns())
	createdOn := qb.TableField{Name: "created_on", Table: record.Meta().GetName()}
	if contains(record.Meta().ReadColumns(), createdOn) {
		updateCols = appendIfMissing(updateCols, createdOn)
	}
	updateOn := qb.TableField{Name: "updated_on", Table: record.Meta().GetName()}
	if contains(record.Meta().ReadColumns(), updateOn) {
		updateCols = appendIfMissing(updateCols, updateOn)
	}
	updateVals := make([]interface{}, len(updateCols))
	for i, col := range updateCols {
		updateVals[i] = columnValues[col.GetName()]
	}

	query := qb.Insert(insertCols...).Values(insertVals...).OnDuplicate(updateCols, updateVals...)
	stmt, values, _ := query.SQL()
	for i, value := range values {
		values[i] = bs.toSQLString(value)
	}
	stmt = strings.Replace(stmt, "?", "%s", -1)
	return fmt.Sprintf(stmt, values...)
}
