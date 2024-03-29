package envset

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
)

type parser struct {
	sliceSeparator string
	envTag         string
	defaultTag     string
	customTypes    map[reflect.Type]func(string) (reflect.Value, error)
	booleans       map[string]bool
}

const (
	defaultEnvTag         = "env"
	defaultDefaultTag     = "default"
	defaultSliceSeparator = ","
)

var defaultBooleans = map[string]bool{
	"1": true, "0": false,
	"t": true, "f": false,
	"true": true, "false": false,
	"y": true, "n": false,
	"yay": true, "nay": false,
	"yes": true, "no": false,
}

// Set accepts a pointer to a struct and zero or more Options
func Set[T any](structPtr *T, options ...Option) error {
	if reflect.TypeOf(structPtr).Elem().Kind() != reflect.Struct {
		// We panic because there is a programmatic error,
		// wrong value passed
		panic(ErrStructPtrExpected)
	}

	return buildParser(options).setStruct(reflect.ValueOf(structPtr).Elem())
}

func buildParser(options []Option) *parser {
	return (&parser{
		sliceSeparator: defaultSliceSeparator,
		envTag:         defaultEnvTag,
		defaultTag:     defaultDefaultTag,
		customTypes:    make(map[reflect.Type]func(string) (reflect.Value, error)),
		booleans:       defaultBooleans,
	}).apply(options)
}

func (p *parser) apply(options []Option) *parser {
	for i := range options {
		options[i](p)
	}

	return p
}

func (p *parser) setStruct(v reflect.Value) error {
	for i := 0; i < v.Type().NumField(); i++ {
		// Skip private fields
		if !v.Type().Field(i).IsExported() {
			continue
		}

		f := v.Field(i)

		// Check if we have a custom type
		if parser, ok := p.customTypes[f.Type()]; ok {
			if err := p.parseType(f, v.Type().Field(i).Tag, parser); err != nil {
				return err
			}
			continue
		}

		// Check if the field is a struct
		if f.Type().Kind() == reflect.Struct {
			if err := p.setStruct(f); err != nil {
				return err
			}
			continue
		}

		// Check if the field is a pointer to a struct
		if f.Kind() == reflect.Pointer && f.Type().Elem().Kind() == reflect.Struct {
			if f.IsNil() {
				f.Set(reflect.New(f.Type().Elem()))
			}

			if err := p.setStruct(f.Elem()); err != nil {
				return err
			}

			continue
		}

		// check if the field already has value
		if !f.IsZero() {
			continue
		}

		// Check if the field is tagged, if not, skip it
		key, ok, optional := p.tagKey(v.Type().Field(i).Tag)
		if !ok {
			continue
		}

		// See if there is an environment variable with name in `key`
		val, ok := os.LookupEnv(key)
		if !ok {
			// Environment var does not exist, check default one
			if val, ok = v.Type().Field(i).Tag.Lookup(p.defaultTag); !ok {
				if optional {
					continue
				}
				// No default, not optional, that's an error
				return NewMissingValueError(key)
			}
		}

		if val == "" && optional {
			continue
		}

		if err := p.setField(f, val, v.Type().Field(i).Tag); err != nil {
			return err
		}
	}

	return nil
}

func (p *parser) setField(f reflect.Value, val string, tags reflect.StructTag) error {
	if f.Kind() == reflect.Pointer && f.IsNil() {
		f.Set(reflect.New(f.Type().Elem()))
		f = f.Elem()
	}

	switch f.Type().Kind() {
	case reflect.Bool:
		return p.parseBool(f, val)
	case reflect.Float32:
		return p.setFloat32(f, tags, val)
	case reflect.Float64:
		return p.setFloat64(f, tags, val)
	case reflect.Int:
		return setInteger[int](f, tags, val)
	case reflect.Uint:
		return setInteger[uint](f, tags, val)
	case reflect.Int8:
		return setInteger[int8](f, tags, val)
	case reflect.Uint8:
		return setInteger[uint8](f, tags, val)
	case reflect.Int16:
		return setInteger[int16](f, tags, val)
	case reflect.Uint16:
		return setInteger[uint16](f, tags, val)
	case reflect.Int32:
		return setInteger[int32](f, tags, val)
	case reflect.Uint32:
		return setInteger[uint32](f, tags, val)
	case reflect.Int64:
		return setInteger[int64](f, tags, val)
	case reflect.Uint64:
		return setInteger[uint64](f, tags, val)
	case reflect.Slice:
		return parseSlice(f, tags, strings.Split(val, p.sliceSeparator))
	case reflect.String:
		return p.setString(f, tags, val)
	default:
		return fmt.Errorf("unsupported type %s of %s", f.Kind(), f.Type())
	}
}

func (p *parser) parseBool(f reflect.Value, val string) error {
	parsed, ok := p.booleans[strings.ToLower(val)]
	if !ok {
		return errors.New("invalid bool value " + val)
	}

	v := reflect.ValueOf(parsed)
	if v.Type() == f.Type() {
		f.Set(v)
	} else if v.CanConvert(f.Type()) {
		f.Set(v.Convert(f.Type()))
	} else {
		return fmt.Errorf("value of type %s is not assignable to field of type %s", v.Type(), f.Type())
	}

	return nil
}

func parseSlice(f reflect.Value, tag reflect.StructTag, values []string) error {
	switch f.Type().Elem().Kind() {
	case reflect.Float32:
		return setSliceOfFloats[float32](values, f)
	case reflect.Float64:
		return setSliceOfFloats[float64](values, f)
	case reflect.Int:
		return setSliceOfIntegers[int](values, f)
	case reflect.Uint:
		return setSliceOfIntegers[uint](values, f)
	case reflect.Int8:
		return setSliceOfIntegers[int8](values, f)
	case reflect.Uint8:
		return setSliceOfIntegers[uint8](values, f)
	case reflect.Int16:
		return setSliceOfIntegers[int16](values, f)
	case reflect.Uint16:
		return setSliceOfIntegers[uint16](values, f)
	case reflect.Int32:
		return setSliceOfIntegers[int32](values, f)
	case reflect.Uint32:
		return setSliceOfIntegers[uint32](values, f)
	case reflect.Int64:
		return setSliceOfIntegers[int64](values, f)
	case reflect.Uint64:
		return setSliceOfIntegers[uint64](values, f)
	case reflect.String:
		if pattern, ok := tag.Lookup("pattern"); ok {
			r, err := regexp.Compile(pattern)
			if err != nil {
				return err
			}

			for i := range values {
				if !r.MatchString(values[i]) {
					return errors.New("value " + values[i] + " does not match pattern " + pattern)
				}
			}
		}

		f.Set(reflect.ValueOf(values))

		return nil
	default:
		return errors.New("unsupported slice elements type: " + f.Type().Elem().Kind().String())
	}
}

func (p *parser) setString(f reflect.Value, tag reflect.StructTag, val string) error {
	if pattern, ok := tag.Lookup("pattern"); ok {
		r, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}

		if !r.MatchString(val) {
			return ErrInvalidValue
		}
	}

	v := reflect.ValueOf(val)
	if v.Type() == f.Type() {
		f.Set(v)
	} else if v.CanConvert(f.Type()) {
		f.Set(v.Convert(f.Type()))
	} else {
		return fmt.Errorf("value of type %s is not assignable to field of type %s", v.Type(), f.Type())
	}

	return nil
}

func (p *parser) parseType(f reflect.Value, tag reflect.StructTag, parser func(string) (reflect.Value, error)) error {
	key, ok, optional := p.tagKey(tag)
	if !ok {
		// No tag, skip it
		return nil
	}

	val, ok := os.LookupEnv(key)
	if !ok {
		// Not set in the environment, check default
		if val, ok = tag.Lookup(p.defaultTag); !ok {
			if optional {
				return nil
			}
			// No default, that's an error
			return NewMissingValueError(key)
		}
	}

	if val == "" {
		if optional {
			return nil
		}

		return NewMissingValueError(key)
	}

	v, err := parser(val)
	if err == nil {
		f.Set(v)
	}

	return err
}

func (p *parser) tagKey(tag reflect.StructTag) (key string, exist, optional bool) {
	if key, exist = tag.Lookup(p.envTag); exist {
		optional = strings.HasSuffix(key, ",omitempty")
		key = strings.TrimSuffix(key, ",omitempty")
	}

	return
}
