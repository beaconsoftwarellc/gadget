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
	// S3GlobalEnvironmentBucketVar is an environment variable for the global environment in S3.
	S3GlobalEnvironmentBucketVar = "S3_GLOBAL_ENV_BUCKET"
	// S3GlobalEnvironmentKeyVar is an environment variable for the global environment in S3.
	S3GlobalEnvironmentKeyVar = "S3_GLOBAL_ENV_KEY"
	// NoSSMEnvVar is the environment variable to set when you do not want to try and pull from SSM.
	NoSSMEnvVar = "NO_SSM_ENV_VARS"
	// SSMEnvironmentVar is the environment variable to set the environment for SSM.
	SSMEnvironmentVar = "SSM_ENVIRONMENT"
	// SSMProjectVar is the environment variable to set the project for SSM.
	SSMProjectVar = "SSM_PROJECT"
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
	envBucket := envVars[S3GlobalEnvironmentBucketVar]
	envItem := []string{envVars[S3GlobalEnvironmentKeyVar]}

	// ssm configuration
	_, noSSM := envVars[NoSSMEnvVar]
	ssm := NewSSM(envVars[SSMEnvironmentVar], envVars[SSMProjectVar])

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typ := val.Type().Field(i)
		envTag, envOptions := stringutil.ParseTag(typ.Tag.Get("env"))
		if stringutil.IsWhiteSpace(envTag) {
			continue
		}

		s3Bucket, s3Item := stringutil.ParseTag(typ.Tag.Get("s3"))
		if stringutil.IsWhiteSpace(s3Bucket) {
			// no s3 configuration for this field, check for it in the global environment
			s3Bucket = envBucket
			s3Item = envItem
		}
		if !stringutil.IsWhiteSpace(s3Bucket) && !noS3 {
			s3Env := bucket.Get(s3Bucket, s3Item[0], envTag, logger)
			if nil != s3Env {
				if typ.Type.Kind() == reflect.Int {
					valueField.SetInt(int64(s3Env.(float64)))
				} else if err := setValueField(valueField, typ, envTag, s3Env.(string)); nil != err {
					return err
				}
				continue
			}
		}

		ssmPath, ssmName := stringutil.ParseTag(typ.Tag.Get("ssm"))
		if len(ssmName) == 0 {
			// if custom ssm tag not specified, use the env tag
			ssmName = []string{envTag}
		}
		if len(ssmName) > 0 && !noSSM {
			ssmEnv := ssm.Get(ssmPath, ssmName[0], logger)
			if !stringutil.IsWhiteSpace(ssmEnv) {
				err := setValueField(valueField, typ, envTag, ssmEnv)
				if nil != err {
					return err
				}
				continue
			}
		}

		env := envVars[envTag]
		if stringutil.IsWhiteSpace(env) {
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
		if stringutil.IsWhiteSpace(envTag) {
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
