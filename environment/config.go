package environment

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

// NoS3EnvVar is the environment variable to set when you so not want to try and pull from S3.
const NoS3EnvVar = "NO_S3_ENV_VARS"

// Process takes a Specification that describes the configuration for the application
// only attributes tagged with `env:""` will be imported from the environment
// Example Specification:
//    type Specification struct {
//        DatabaseURL string `env:"DATABASE_URL"`
//        ServiceID   string `env:"SERVICE_ID,optional"`
//    }
// Supported options: optional
func Process(config interface{}) error {
	return ProcessMap(config, GetEnvMap())
}

// ProcessMap is the same as Process except that the environment variable map is supplied instead of retrieved
func ProcessMap(config interface{}, envVars map[string]string) error {
	val := reflect.ValueOf(config)

	if val.Kind() != reflect.Ptr {
		return NewInvalidSpecificationError()
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return NewInvalidSpecificationError()
	}

	bucket := NewBucket()
	_, noS3 := envVars[NoS3EnvVar]
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typ := val.Type().Field(i)
		envTag, envOptions := stringutil.ParseTag(typ.Tag.Get("env"))
		if "" == envTag {
			continue
		}

		s3Bucket, s3Item := stringutil.ParseTag(typ.Tag.Get("s3"))
		if "" != s3Bucket && !noS3 {
			s3Env := bucket.Get(s3Bucket, s3Item[0], envTag)
			if nil != s3Env {
				switch t := typ.Type.Kind(); t {
				case reflect.String:
					valueField.SetString(s3Env.(string))
				case reflect.Int:
					valueField.SetInt(int64(s3Env.(float64)))
				default:
					return UnsupportedDataTypeError{Type: t, Field: typ.Name}
				}
				continue
			}
		}

		env := envVars[envTag]
		if "" == env {
			if !envOptions.Contains("optional") {
				return MissingEnvironmentVariableError{Tag: envTag, Field: typ.Name}
			}
			continue
		}

		switch valueField.Interface().(type) {
		case string:
			valueField.SetString(env)
		case int:
			parsed, err := strconv.Atoi(env)
			if err != nil {
				return errors.New("%s while converting %s", err.Error(), envTag)
			}
			valueField.SetInt(int64(parsed))
		case time.Duration:
			parsed, err := time.ParseDuration(env)
			if err != nil {
				return errors.New("%s while converting %s", err.Error(), envTag)
			}
			valueField.SetInt(int64(parsed))
		default:
			return UnsupportedDataTypeError{Type: typ.Type.Kind(), Field: typ.Name}
		}
	}
	return nil
}

// Push the passed specification object onto the environment, note that these changes
// will not live past the life of this process.
func Push(config interface{}) error {
	val := reflect.ValueOf(config)

	if val.Kind() != reflect.Ptr {
		return NewInvalidSpecificationError()
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return NewInvalidSpecificationError()
	}
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typ := val.Type().Field(i)
		envTag, _ := stringutil.ParseTag(typ.Tag.Get("env"))
		if "" == envTag {
			continue
		}
		var value string

		switch t := valueField.Interface().(type) {
		case string:
			value = t
		case int:
			value = strconv.Itoa(t)
		case time.Duration:
			value = t.String()
		default:
			return UnsupportedDataTypeError{Type: typ.Type.Kind(), Field: typ.Name}
		}
		os.Setenv(envTag, value)
	}
	return nil
}

// GetEnvMap returns a map of all environment variables to their values
func GetEnvMap() map[string]string {
	raw := os.Environ() //format "key=val"
	filtered := make(map[string]string, len(raw))
	for _, keyval := range raw {
		parts := strings.SplitN(keyval, "=", 2)
		filtered[parts[0]] = parts[1]
	}
	return filtered
}
