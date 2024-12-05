package record

import (
	"reflect"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
)

// Record defines a database enabled record
type Record interface {
	// Initialize sets any expected values on a Record during Create
	Initialize()
	// PrimaryKey returns the value of the primary key of the Record
	PrimaryKey() PrimaryKeyValue
	// Key returns the name of the primary key of the Record
	Key() string
	// Meta returns the meta object for this Record
	Meta() qb.Table
}

// DefaultRecord implements the Key() as "id"
type DefaultRecord struct{}

// Key returns "ID" as the default Primary Key
func (record *DefaultRecord) Key() string {
	return "id"
}

// PrimaryKeyValue limits keys to string or int
type PrimaryKeyValue struct {
	intPK int
	strPK string
	isInt bool
}

// NewPrimaryKey returns a populated PrimaryKeyValue
func NewPrimaryKey(value interface{}) (pk PrimaryKeyValue) {
	switch t := reflect.TypeOf(value).Kind(); t {
	case reflect.String:
		pk.strPK = value.(string)
	case reflect.Int:
		pk.intPK = value.(int)
		pk.isInt = true
	}
	return
}

// Value returns the string or integer value for a Record
func (pk PrimaryKeyValue) Value() interface{} {
	if pk.isInt {
		return pk.intPK
	}
	return pk.strPK
}

// IsInteger returns true if the Primary Key is an integer
func (pk PrimaryKeyValue) IsInteger() bool {
	return pk.isInt
}
