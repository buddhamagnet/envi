package envi

import (
	"reflect"
	"errors"
	"strings"
	"os"
	"strconv"
	"fmt"
)

var (
	//Errors
	ErrNotAPtrStruct = errors.New("Expected a pointer to a Struct")
	UnsupportedType = errors.New("Unsupported Type")

	//Slices
	sIntS = reflect.TypeOf([]int(nil))
	SInt64S = reflect.TypeOf([]int64(nil))
	sStringS = reflect.TypeOf([]string(nil))
	sBoolS = reflect.TypeOf([]bool(nil))
	sFloat32S = reflect.TypeOf([]float32(nil))
	sFloat64S = reflect.TypeOf([]float64(nil))
)

const (
	// Keys
	Blank = ""
	Env = "env"
	EnvDefault = "envDefault"
	EnvSeparator = "envSeparator"

	// Options Support
	Required = "required"
)

// Public functions
func Parse(val interface{}) error {

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
		if err := setValue(val.Field(i), refType.Field(i), value); err != nil {
			errs = append(errs, err.Error())
			continue
		}
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
		err error
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

func setValue(field reflect.Value, sf reflect.StructField, val string) error {

	// Case for Type
	// With field.Kind() get kind represents the specific kind of type that a Type represents.
	switch field.Kind() {
	case reflect.Slice:
		separator := sf.Tag.Get(EnvSeparator)
		return doHandleSlice(field, val, separator)
	case reflect.String:
		field.SetString(val)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	case reflect.Int:
		intValue, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return err
		}
		field.SetInt(intValue)
	case reflect.Int64:
		int64Value, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(int64Value)
	case reflect.Uint:
		uintValue, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case reflect.Float32:
		float32Value, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return err
		}
		field.SetFloat(float32Value)
	case reflect.Float64:
		float64Value, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(float64Value))
	default:
		return UnsupportedType
	}
	return nil
}

func doHandleSlice(field reflect.Value, value, separator string) error {
	if separator == Blank {
		separator = ","
	}

	splitData := strings.Split(value, separator)

	switch field.Type() {
	case sStringS:
		field.Set(reflect.ValueOf(splitData))
	case sBoolS:
		data, err := doParseBoolS(splitData)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(data))
	case sIntS:
		data, err := doParseIntS(splitData)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(data))
	case SInt64S:
		data, err := doParseInt64S(splitData)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(data))

	case sFloat32S:
		data, err := doParseFloat32S(splitData)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(data))
	case sFloat64S:
		data, err := doParseFloat64S(splitData)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(data))
	default:
		return errors.New(fmt.Sprintf("Unsupported Slice Type %s.", field.Type()))
	}
	return nil
}

func doParseBoolS(d []string) ([]bool, error) {
	var boolSlice []bool

	for _, v := range d {
		boolValue, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}

		boolSlice = append(boolSlice, boolValue)
	}
	return boolSlice, nil
}

func doParseIntS(d []string) ([]int, error) {
	var intSlice []int

	for _, v := range d {
		intValue, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, err
		}
		intSlice = append(intSlice, int(intValue))
	}
	return intSlice, nil
}

func doParseFloat32S(d []string) ([]float32, error) {
	var float32Slice []float32

	for _, v := range d {
		data, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return nil, err
		}
		float32Slice = append(float32Slice, float32(data))
	}
	return float32Slice, nil
}

func doParseInt64S(d []string) ([]int64, error) {
	var int64Slice []int64

	for _, v := range d {
		intValue, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		int64Slice = append(int64Slice, int64(intValue))
	}
	return int64Slice, nil
}

func doParseFloat64S(d []string) ([]float64, error) {
	var float64Slice []float64

	for _, v := range d {
		data, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		float64Slice = append(float64Slice, float64(data))
	}
	return float64Slice, nil
}