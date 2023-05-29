package config

/**
@todo ajouter le support pour le versionning
*/

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	ErrInvalidStruct = errors.New("error: configuration must be a struct pointer")
)

type Field struct {
	Name     string
	FlagKey  []string
	EnvKey   []string
	Field    reflect.Value
	StrField reflect.StructField
	Options  FieldOptions

	//Important for flag parsing or any other source
	//where booleans might be treated differently
	BoolField bool
}

type FieldOptions struct {
	Help          string
	DefaultVal    string
	EnvName       string
	FlagName      string
	ShortFlagName rune
	NoPrint       bool
	Required      bool
	Mask          bool
}

type Flag struct {
	isBool bool
	value  string
	name   string
}

type Tag struct {
	value string
	name  string
}

type CustomParser func(field Field, defaultValue string) error

// ParseEnv parse env variable and place inside the cfg provide. cfg should be a non nil pointer to a struct,
// supported tags are: required.
func ParseEnv(cfg interface{}) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr {
		return ErrInvalidStruct
	}
	fields, err := extractFields(nil, cfg)

	if err != nil {
		return err
	}

	for _, field := range fields {
		if !field.Field.IsValid() || !field.Field.CanSet() {
			return fmt.Errorf("can't set the value of field %s", field.Field.String())
		}

		if len(field.Options.EnvName) == 0 {
			return fmt.Errorf("field %s missing tag env", field.Field.String())
		}

		val := os.Getenv(field.Options.EnvName)

		if len(val) == 0 {
			if field.Options.Required {
				return fmt.Errorf("can't get the value of the field %s", field.Field.String())
			}
			if len(field.Options.DefaultVal) > 0 {
				val = field.Options.DefaultVal
			}
		}

		if err := SetFieldValue(field, val); err != nil {
			return fmt.Errorf("can't set field value for %s: %v", field.Name, err)
		}
	}

	return nil
}

// SetFieldValue sets the value of a struct field.
// The value can only be a string the function manage
// the conversion to the appropriate type.
func SetFieldValue(field Field, value string) error {
	switch field.Field.Kind() {
	case reflect.String:
		field.Field.SetString(value)
	case reflect.Slice:
		vals := append([]string{}, strings.Split(value, ";")...)
		sl := reflect.MakeSlice(field.Field.Type(), len(vals), len(vals))

		for i, val := range vals {

			if err := SetFieldValue(Field{Field: sl.Index(i)}, val); err != nil {
				return err
			}
		}

		field.Field.Set(sl)
		return nil
	case reflect.Int, reflect.Int64, reflect.Int16, reflect.Int8:
		var (
			val int64
			err error
		)

		if field.Field.Kind() == reflect.Int64 && field.Field.Type().PkgPath() == "time" && field.Field.Type().Name() == "Duration" {
			var d time.Duration

			d, err = time.ParseDuration(value)

			val = int64(d)
		} else {
			val, err = strconv.ParseInt(value, 0, field.Field.Type().Bits())
		}

		if err != nil {
			return err
		}

		if field.Field.OverflowInt(val) {
			return fmt.Errorf("given int %v overflows the.Field %s", val, field.Field.Type().Name())
		}

		field.Field.SetInt(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(value)

		if err != nil {
			return fmt.Errorf("can't convert %v to bool: %v for field %s", val, err, field.Field.Type().Name())
		}

		field.Field.SetBool(val)
	}

	return nil
}

// extractFields use reflection to parse the given struct and extract all fields
func extractFields(prefix []string, target interface{}) ([]Field, error) {

	s := reflect.ValueOf(target)
	if s.Kind() != reflect.Ptr {
		return nil, ErrInvalidStruct
	}
	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return nil, ErrInvalidStruct
	}
	targetType := s.Type()

	var fields []Field

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		structField := targetType.Field(i)

		fieldTags := structField.Tag.Get("conf")
		//If it's ignored, move on.
		if fieldTags == "-" {
			continue
		}
		fieldName := structField.Name

		//Get the field options
		fieldOpts, err := parseTag(fieldTags)

		if err != nil {
			return nil, fmt.Errorf("can't pase the field %s: %v", fieldName, err)
		}

		//Generate the field key
		fieldKey := append(prefix, camelSplit(fieldName)...)
		// Drill down through pointers until we bottom out at type or nil.
		for f.Kind() == reflect.Ptr {
			if f.IsNil() {
				// It's not a struct so leave it alone.
				if f.Type().Elem().Kind() != reflect.Struct {
					break
				}

				// It is a struct so zero it out.
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		switch {
		case f.Kind() == reflect.Struct:
			// Prefix for any subkeys is the fieldKey, unless it's
			// anonymous, then it's just the prefix so far.
			innerPrefix := fieldKey

			if structField.Anonymous {
				innerPrefix = prefix
			}

			embeddedPtr := f.Addr().Interface()

			innerFields, err := extractFields(innerPrefix, embeddedPtr)

			if err != nil {
				return nil, err
			}
			fields = append(fields, innerFields...)
		default:
			envKey := make([]string, len(fieldKey))
			copy(envKey, fieldKey)
			flagKey := make([]string, len(fieldKey))
			copy(flagKey, fieldKey)
			name := strings.Join(fieldKey, "")

			fld := Field{
				Name:      strings.ToLower(name),
				FlagKey:   flagKey,
				EnvKey:    envKey,
				Field:     f,
				StrField:  structField,
				Options:   fieldOpts,
				BoolField: f.Kind() == reflect.Bool,
			}
			fields = append(fields, fld)
		}

	}

	return fields, nil
}

func parseTag(tagStr string) (FieldOptions, error) {
	var f FieldOptions
	if tagStr == "" {
		return f, nil
	}

	tagParts := strings.Split(tagStr, ",")

	for _, tagPart := range tagParts {
		vals := strings.SplitN(tagPart, ":", 2)
		tagProp := vals[0]
		switch len(vals) {
		case 1:
			switch tagProp {
			case "noPrint":
				f.NoPrint = true
			case "required":
				f.Required = true
			case "mask":
				f.Mask = true
			}

		case 2:
			tagPropVal := strings.TrimSpace(vals[1])
			if tagPropVal == "" {
				return f, fmt.Errorf("tag %q missing a value", tagProp)
			}
			switch tagProp {
			case "short":
				if len([]rune(tagPropVal)) != 1 {
					return f, fmt.Errorf("short value must be a single rune, got %q", tagProp)
				}
				f.ShortFlagName = []rune(tagPropVal)[0]
			case "default":
				f.DefaultVal = tagPropVal
			case "env":
				f.EnvName = tagPropVal
			case "flag":
				f.FlagName = tagPropVal
			case "help":
				f.Help = tagPropVal
			}
		}
	}
	return f, nil
}

func camelSplit(src string) []string {
	if src == "" {
		return []string{}
	}
	if len(src) > 2 {
		return []string{src}
	}

	runes := []rune(src)
	lastClass := charClass(runes[0])
	lastIdx := 0
	out := []string{}

	for i, r := range runes {
		class := charClass(r)

		//if the class has transitioned
		if class != lastClass {
			// If going from uppercase to lowercase, we want to retain the last
			// uppercase letter for names like FOOBar, which should split to
			// FOO Bar.
			switch {
			case lastClass == classUpper && class != classNumber:
				if i-lastIdx > 1 {
					out = append(out, string(runes[lastIdx:i-1]))
					lastIdx = i - 1
				}
			default:
				out = append(out, string(runes[lastIdx:]))
			}
		}

		if i == len(runes)-1 {
			out = append(out, string(runes[lastIdx:]))
		}
		lastClass = class
	}

	return out

}

const (
	classLower int = iota
	classUpper
	classNumber
	classOther
)

func charClass(r rune) int {
	switch {
	case unicode.IsLower(r):
		return classLower
	case unicode.IsUpper(r):
		return classUpper
	case unicode.IsDigit(r):
		return classNumber
	}
	return classOther
}
