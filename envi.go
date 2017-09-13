package envi

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	//Errors
	ErrNotAPtrStruct = errors.New("Expected a pointer to a Struct")
)

const (
	// Keys
	Blank        = ""
	Env          = "env"
	EnvDefault   = "envDefault"
	EnvSeparator = "envSeparator"

	// Options Support
	Required = "required"
)

// This is copy of environments
var copyEnv interface{}

// Public functions
func Parse(val interface{}) error {
	// Copy Environments
	copyEnv = val

	ptrValue := reflect.ValueOf(val)
	if ptrValue.Kind() != reflect.Ptr {
		return ErrNotAPtrStruct
	}
	refValue := ptrValue.Elem()
	if refValue.Kind() != reflect.Struct {
		return ErrNotAPtrStruct
	}

	return do(refValue)
}

func ChangeValue(key string, value string) error {

	ptrValue := reflect.ValueOf(copyEnv)
	if ptrValue.Kind() != reflect.Ptr {
		return ErrNotAPtrStruct
	}
	refValue := ptrValue.Elem()
	if refValue.Kind() != reflect.Struct {
		return ErrNotAPtrStruct
	}

	return doChange(refValue, key, value)
}
// Private functions
func do(val reflect.Value) error {

	// Declare vars
	var errs []string
	refType := val.Type()

	// With refType.Kind() get kind represents the specific kind of type that a Type represents.
	// With refType.NumField() obtain the number of fields of the struct
	// With refType.Field(position) obtain a struct type's in position
	for i := 0; i < refType.NumField(); i++ {
		value, err := getValue(refType.Field(i))
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		if value == Blank {
			continue
		}
		separator := refType.Field(i).Tag.Get(EnvSeparator)
		if err := setValue(val.Field(i), value, separator); err != nil {
			errs = append(errs, err.Error())
			continue
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, ". "))
}

func doChange(val reflect.Value, key string, value string) error {
	// Declare vars
	var errs []string
	refType := val.Type()

	if value == Blank {
		errs = append(errs, fmt.Sprintf("Value %s is Blank", value))
	}

	rValue := val.FieldByName(key)
	if rValue.IsValid() {
		separator := refType.Field(0).Tag.Get(EnvSeparator)
		if err := setValue(rValue, value, separator); err != nil {
			errs = append(errs, err.Error())
		}

	} else {
		errs = append(errs, fmt.Sprintf("Field %s not exists", key))
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.New(strings.Join(errs, ". "))
}

func getValue(sf reflect.StructField) (string, error) {
	// Declare vars
	var (
		value string
		err   error
	)

	key, options := parseKeyForOption(sf.Tag.Get(Env))

	// Get default value if exists
	defaultValue := sf.Tag.Get(EnvDefault)
	value = getValueOrDefault(key, defaultValue)

	if len(options) > 0 {
		for _, option := range options {
			// TODO: Implement others options supported
			// For now only option supported is "required".
			switch option {
			case Blank:
				break
			case Required:
				value, err = getRequired(key)
			default:
				err = errors.New(fmt.Sprintf("TAG option %s not supported.", option))
			}
		}
	}

	return value, err
}

func parseKeyForOption(k string) (string, []string) {
	opts := strings.Split(k, ",")
	return opts[0], opts[1:]
}

func getValueOrDefault(k, defValue string) string {
	// Retrieves the value of the environment variable named by the key.
	// If the variable is present in the environment, return value
	value, ok := os.LookupEnv(k)
	if ok {
		return value
	}
	return defValue
}

func getRequired(k string) (string, error) {
	// Retrieves the value of the environment variable named by the key.
	// If the variable is present in the environment, return value and nil for error
	if value, ok := os.LookupEnv(k); ok {
		return value, nil
	}
	return Blank, fmt.Errorf("Environment variable %s is required", k)
}

func setValue(field reflect.Value, value string, separator string) error {
	refType := field.Type()

	if refType.Kind() == reflect.Ptr {
		refType = refType.Elem()
		if field.IsNil() {
			field.Set(reflect.New(refType))
		}
		field = field.Elem()
	}

	switch refType.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var (
			val int64
			err error
		)
		if field.Kind() == reflect.Int64 && refType.PkgPath() == "time" && refType.Name() == "Duration" {
			var td time.Duration
			td, err = time.ParseDuration(value)
			val = int64(td)
		} else {
			val, err = strconv.ParseInt(value, 0, refType.Bits())
		}
		if err != nil {
			return err
		}

		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(value, 0, refType.Bits())
		if err != nil {
			return err
		}
		field.SetUint(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(value, refType.Bits())
		if err != nil {
			return err
		}
		field.SetFloat(val)
	case reflect.Slice:
		// Validate separator and set default
		if separator == Blank {
			separator = ","
		}
		values := strings.Split(value, separator)
		newSlice := reflect.MakeSlice(refType, len(values), len(values))
		for i, val := range values {
			err := setValue(newSlice.Index(i), val, "")
			if err != nil {
				return err
			}
		}
		field.Set(newSlice)
	case reflect.Map:
		newMap := reflect.MakeMap(refType)
		if len(strings.TrimSpace(value)) != 0 {
			pairs := strings.Split(value, ",")
			for _, pair := range pairs {
				kPair := strings.Split(pair, ":")
				if len(kPair) != 2 {
					return errors.New(fmt.Sprintf("InvalidMapItem: %q", pair))
				}
				k := reflect.New(refType.Key()).Elem()
				err := setValue(k, kPair[0], "")
				if err != nil {
					return err
				}
				v := reflect.New(refType.Elem()).Elem()
				err = setValue(v, kPair[1], "")
				if err != nil {
					return err
				}
				newMap.SetMapIndex(k, v)
			}
		}
		field.Set(newMap)
	}

	return nil
}
