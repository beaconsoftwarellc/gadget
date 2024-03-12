package environment

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

const (
	// NoS3EnvVar is the environment variable to set when you do not want to try and pull from S3.
	NoS3EnvVar = "NO_S3_ENV_VARS"
	// NoSSMEnvVar is the environment variable to set when you do not want to try and pull from SSM.
	NoSSMEnvVar = "NO_SSM_ENV_VARS"
	// SSMEnvironmentVar is the environment variable to set the environment for SSM.
	SSMEnvironmentVar = "SSM_ENVIRONMENT"
)

// Process takes a Specification that describes the configuration for the application
// only attributes tagged with `env:""` will be imported from the environment
// Example Specification:
//
//	type Specification struct {
//	    DatabaseURL string `env:"DATABASE_URL"`
//	    ServiceID   string `env:"SERVICE_ID,optional"`
//	}
//
// Supported options: optional
func Process(config interface{}, logger log.Logger) error {
	return ProcessMap(config, GetEnvMap(), logger)
}

// ProcessMap is the same as Process except that the environment variable map is supplied instead of retrieved
func ProcessMap(config interface{}, envVars map[string]string, logger log.Logger) error {
	val := reflect.ValueOf(config)

	if val.Kind() != reflect.Ptr {
		return NewInvalidSpecificationError()
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return NewInvalidSpecificationError()
	}

	// s3 configuration
	bucket := NewBucket()
	_, noS3 := envVars[NoS3EnvVar]

	// ssm configuration
	_, noSSM := envVars[NoSSMEnvVar]
	ssmEnv := envVars[SSMEnvironmentVar]
	noSSM = noSSM || ssmEnv == ""
	ssm := NewSSM(ssmEnv)

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typ := val.Type().Field(i)
		envTag, envOptions := stringutil.ParseTag(typ.Tag.Get("env"))
		if envTag == "" {
			continue
		}

		s3Bucket, s3Item := stringutil.ParseTag(typ.Tag.Get("s3"))
		if s3Bucket != "" && !noS3 {
			s3Env := bucket.Get(s3Bucket, s3Item[0], envTag, logger)
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

		ssmName, _ := stringutil.ParseTag(typ.Tag.Get("ssm"))
		if ssmName != "" && !noSSM {
			ssmEnv := ssm.Get(ssmName, logger)
			if ssmEnv != "" {
				err := setValueField(valueField, typ, envTag, ssmEnv)
				if nil != err {
					return err
				}
				continue
			}
		}

		env := envVars[envTag]
		if env == "" {
			if !envOptions.Contains("optional") {
				return MissingEnvironmentVariableError{Tag: envTag, Field: typ.Name}
			}
			continue
		}

		err := setValueField(valueField, typ, envTag, env)
		if nil != err {
			return err
		}
	}
	return nil
}

func setValueField(valueField reflect.Value, structField reflect.StructField,
	envTag, env string) error {
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
		return UnsupportedDataTypeError{Type: structField.Type.Kind(), Field: structField.Name}
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
		if envTag == "" {
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
