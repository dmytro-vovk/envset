package envset

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type parser struct {
	sliceSeparator string
	envTag         string
	defaultTag     string
	customTypes    map[string]func(string) (reflect.Value, error)
}

const (
	defaultEnvTag         = "env"
	defaultDefaultTag     = "default"
	defaultSliceSeparator = ","
)

func Set(s any, options ...Option) error {
	t := reflect.TypeOf(s)

	if t.Kind() != reflect.Pointer {
		return ErrStructExpected
	}

	if t.Elem().Kind() != reflect.Struct {
		return ErrStructExpected
	}

	return buildParser(options...).setStruct(reflect.ValueOf(s).Elem())
}

func buildParser(options ...Option) *parser {
	p := parser{
		sliceSeparator: defaultSliceSeparator,
		envTag:         defaultEnvTag,
		defaultTag:     defaultDefaultTag,
		customTypes:    make(map[string]func(string) (reflect.Value, error)),
	}

	for i := range options {
		options[i](&p)
	}

	return &p
}

func (p *parser) setStruct(v reflect.Value) error {
	for i := 0; i < v.Type().NumField(); i++ {
		// Skip private fields
		if !v.Type().Field(i).IsExported() {
			continue
		}

		f := v.Field(i)

		// Check if we have a custom type
		if _, ok := p.customTypes[f.Type().String()]; ok {
			if err := p.parseType(f, v.Type().Field(i).Tag); err != nil {
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

		// Check if the field is tagged, if not, skip it
		val, ok := p.tagValue(v.Type().Field(i).Tag)
		if !ok {
			continue
		}

		if val == "" {
			return errors.New("value required, but not set")
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
	case reflect.Float64:
		return p.parseFloat(f, tags, val)
	case reflect.Int:
		return p.parseInt(f, tags, val)
	case reflect.Slice:
		return p.parseSlice(f, tags, val)
	case reflect.String:
		return p.parseString(f, tags, val)
	default:
		return fmt.Errorf("unsupported type %s of %s", f.Kind(), f.Type())
	}
}

func (p *parser) parseBool(f reflect.Value, val string) error {
	parsed, ok := map[string]bool{
		"true": true, "false": false,
		"yes": true, "no": false,
		"1": true, "0": false,
	}[strings.ToLower(val)]

	if !ok {
		return errors.New("invalid bool value " + val)
	}

	f.Set(reflect.ValueOf(parsed))

	return nil
}

func (p *parser) parseSlice(f reflect.Value, tags reflect.StructTag, val string) error {
	switch f.Type().Elem().Kind() {
	case reflect.Float64:
		var values []float64

		for _, v := range strings.Split(val, p.sliceSeparator) {
			i, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return err
			}

			values = append(values, i)
		}

		f.Set(reflect.ValueOf(values))
	case reflect.Int:
		var values []int

		for _, v := range strings.Split(val, p.sliceSeparator) {
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}

			values = append(values, i)
		}

		f.Set(reflect.ValueOf(values))
	case reflect.String:
		values := strings.Split(val, p.sliceSeparator)

		if pattern, ok := tags.Lookup("pattern"); ok {
			r, err := regexp.Compile(pattern)
			if err != nil {
				return err
			}

			for i := range values {
				if !r.MatchString(values[i]) {
					return errors.New("invalid value " + values[i])
				}
			}
		}

		f.Set(reflect.ValueOf(values))
	default:
		return errors.New("unsupported slice elements type: " + f.Type().Elem().Kind().String())
	}

	return nil
}

func (p *parser) parseInt(f reflect.Value, tags reflect.StructTag, val string) error {
	i, err := strconv.Atoi(val)
	if err != nil {
		return err
	}

	if min, ok := tags.Lookup("min"); ok {
		if m, err := strconv.Atoi(min); err != nil {
			return err
		} else if i < m {
			return errors.New("value " + val + " is less tha minimal value " + min)
		}
	}

	if max, ok := tags.Lookup("max"); ok {
		if m, err := strconv.Atoi(max); err != nil {
			return err
		} else if i > m {
			return errors.New("value " + val + " is greater than maximum value " + max)
		}
	}

	f.Set(reflect.ValueOf(i))

	return nil
}

func (p *parser) parseFloat(f reflect.Value, tags reflect.StructTag, val string) error {
	i, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return err
	}

	if min, ok := tags.Lookup("min"); ok {
		if m, err := strconv.ParseFloat(min, 64); err != nil {
			return err
		} else if i < m {
			return errors.New("value " + val + " is less tha minimal value " + min)
		}
	}

	if max, ok := tags.Lookup("max"); ok {
		if m, err := strconv.ParseFloat(max, 64); err != nil {
			return err
		} else if i > m {
			return errors.New("value " + val + " is greater than maximum value " + max)
		}
	}

	f.Set(reflect.ValueOf(i))

	return nil
}

func (p *parser) parseString(f reflect.Value, tags reflect.StructTag, val string) error {
	if pattern, ok := tags.Lookup("pattern"); ok {
		r, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}

		if !r.MatchString(val) {
			return ErrInvalidValue
		}
	}

	f.Set(reflect.ValueOf(val))

	return nil
}

func (p *parser) parseType(f reflect.Value, tags reflect.StructTag) error {
	val, ok := p.tagValue(tags)
	if !ok {
		return nil
	}

	if val == "" {
		return errors.New("value required, but not set")
	}

	if fn, ok := p.customTypes[f.Type().String()]; ok {
		v, err := fn(val)
		if err != nil {
			return err
		}

		f.Set(v)

		return nil
	}

	return errors.New("unsupported field type " + f.Type().String())
}

func (p *parser) tagValue(tags reflect.StructTag) (string, bool) {
	env, ok := tags.Lookup(p.envTag)
	if !ok {
		return "", false
	}

	omitEmpty := strings.HasSuffix(env, ",omitempty")
	if omitEmpty {
		env = strings.TrimSuffix(env, ",omitempty")
	}

	val, ok := os.LookupEnv(env)
	if !ok {
		val, ok = tags.Lookup(p.defaultTag)
		if !ok {
			if omitEmpty {
				return "", false
			}
			return "", true
		}
	}

	return val, true
}
