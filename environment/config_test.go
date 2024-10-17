package environment

import (
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	assert1 "github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
)

type specification struct {
	StringField         string        `env:"STRING_FIELD" s3:"bar,foo"`
	IntField            int           `env:"INT_FIELD,junk" s3:"foo,bar"`
	BoolField           bool          `env:"BOOL_FIELD" s3:"foo,bar"`
	OptionalField       string        `env:"OPTIONAL_FIELD,optional,junk"`
	Interval            time.Duration `env:"INTERVAL,optional"`
	NotEnvironmentField string
}

type unsupportedTypeSpecification struct {
	Float64Field float64 `env:"FLOAT64_FIELD" s3:"invalid,type"`
}

func TestPush(t *testing.T) {
	assert := assert1.New(t)
	os.Clearenv()

	spec := &specification{
		StringField:   generator.String(20),
		IntField:      10,
		BoolField:     true,
		OptionalField: generator.String(30),
		Interval:      time.Duration(int64(generator.Int16())),
	}

	assert.NoError(Push(spec))
	assert.Equal(os.Getenv("STRING_FIELD"), spec.StringField)
	assert.Equal(os.Getenv("INT_FIELD"), "10")
	assert.Equal(os.Getenv("BOOL_FIELD"), "true")
	assert.Equal(os.Getenv("OPTIONAL_FIELD"), spec.OptionalField)
	assert.Equal(os.Getenv("INTERVAL"), spec.Interval.String())
}

func TestValidConfig(t *testing.T) {
	assert := assert1.New(t)
	os.Clearenv()

	expectedStringField := "life, the universe and everything"
	os.Setenv("STRING_FIELD", expectedStringField)

	expectedIntField := 42
	os.Setenv("INT_FIELD", strconv.Itoa(expectedIntField))

	interval := time.Duration(int64(generator.Int16()))
	os.Setenv("INTERVAL", interval.String())

	expectedBoolField := true
	os.Setenv("BOOL_FIELD", strconv.FormatBool(expectedBoolField))

	expectedNotEnvironmentField := "How many roads must a man walk down?"
	config := &specification{NotEnvironmentField: expectedNotEnvironmentField}
	err := Process(config, log.NewStackLogger())

	assert.NoError(err)
	assert.Equal(expectedStringField, config.StringField)
	assert.Equal(expectedIntField, config.IntField)
	assert.Equal(expectedBoolField, config.BoolField)
	assert.Equal("", config.OptionalField)
	assert.Equal(expectedNotEnvironmentField, config.NotEnvironmentField)
	assert.Equal(interval, config.Interval)
}

func TestProcessNonPointerFails(t *testing.T) {
	assert := assert1.New(t)
	os.Clearenv()

	expectedStringField := "life, the universe and everything"
	os.Setenv("STRING_FIELD", expectedStringField)

	expectedIntField := 42
	os.Setenv("INT_FIELD", strconv.Itoa(expectedIntField))

	expectedNotEnvironmentField := "How many roads must a man walk down?"
	config := specification{NotEnvironmentField: expectedNotEnvironmentField}
	err := Process(config, log.NewStackLogger())

	assert.EqualError(err, NewInvalidSpecificationError().Error())
	assert.Equal(specification{NotEnvironmentField: expectedNotEnvironmentField}, config)
}

func TestMissingEnviroment(t *testing.T) {
	assert := assert1.New(t)
	os.Clearenv()

	expectedStringField := "life, the universe and everything"
	os.Setenv("STRING_FIELD", expectedStringField)

	expectedNotEnvironmentField := "How many roads must a man walk down?"
	config := &specification{NotEnvironmentField: expectedNotEnvironmentField}
	err := Process(config, log.NewStackLogger())

	assert.EqualError(err, MissingEnvironmentVariableError{Tag: "INT_FIELD", Field: "IntField"}.Error())
	assert.Equal(expectedStringField, config.StringField)
	assert.Equal(0, config.IntField)
	assert.Equal(expectedNotEnvironmentField, config.NotEnvironmentField)
}

func TestNotImplementedType(t *testing.T) {
	assert := assert1.New(t)
	os.Clearenv()

	os.Setenv("FLOAT64_FIELD", "20.12")

	config := &unsupportedTypeSpecification{}
	err := Process(config, log.NewStackLogger())

	assert.EqualError(err, UnsupportedDataTypeError{Type: reflect.Float64, Field: "Float64Field"}.Error())
	assert.Equal(&unsupportedTypeSpecification{}, config)
}

func TestInvalidConfigValue(t *testing.T) {
	assert := assert1.New(t)
	os.Clearenv()

	expectedStringField := "life, the universe and everything"
	os.Setenv("STRING_FIELD", expectedStringField)

	os.Setenv("INT_FIELD", "j")

	expectedNotEnvironmentField := "How many roads must a man walk down?"
	config := &specification{NotEnvironmentField: expectedNotEnvironmentField}
	err := Process(config, log.NewStackLogger())

	assert.Error(err)
	assert.Equal("strconv.Atoi: parsing \"j\": invalid syntax while converting INT_FIELD", err.Error())
	assert.Equal(expectedStringField, config.StringField)
	assert.Equal(0, config.IntField)
	assert.Equal("", config.OptionalField)
	assert.Equal(expectedNotEnvironmentField, config.NotEnvironmentField)
}

func TestNonStructProcessed(t *testing.T) {
	assert := assert1.New(t)
	os.Clearenv()

	config := "42"
	err := Process(&config, log.NewStackLogger())

	assert.EqualError(err, NewInvalidSpecificationError().Error())
}
