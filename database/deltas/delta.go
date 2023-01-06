package deltas

import (
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/database/record"
)

// Delta represents a set of changes that are to be applied to the database as an atomic unit
type Delta struct {
	ID     int
	Name   string
	Script string
}

// DeltaRecord is the database representation of delta script and indicates that it has been
// executed on the database.
type DeltaRecord struct {
	record.DefaultRecord
	ID       int       `db:"id"`
	Name     string    `db:"name"`
	Created  time.Time `db:"created,read_only"`
	Modified time.Time `db:"modified,read_only"`
}

// Initialize the delta record with an id
func (dbm *DeltaRecord) Initialize() {}

// PrimaryKey of this record
func (dbm *DeltaRecord) PrimaryKey() record.PrimaryKeyValue {
	return record.NewPrimaryKey(dbm.ID)
}

// Key field name
func (dbm *DeltaRecord) Key() string {
	return "id"
}

// Meta object for this record
func (dbm *DeltaRecord) Meta() qb.Table {
	return DeltaMeta
}
