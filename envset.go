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

	if t.Kind() != reflect.Pointer || t.Elem().Kind() != reflect.Struct {
		// We panic because there is a programmatic error,
		// wrong value passed
		panic(ErrStructPtrExpected)
	}

	return buildParser(options).setStruct(reflect.ValueOf(s).Elem())
}

func buildParser(options []Option) *parser {
	return (&parser{
		sliceSeparator: defaultSliceSeparator,
		envTag:         defaultEnvTag,
		defaultTag:     defaultDefaultTag,
		customTypes:    make(map[string]func(string) (reflect.Value, error)),
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
		if parser, ok := p.customTypes[f.Type().String()]; ok {
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
	case reflect.Float32:
		return p.parseFloat32(f, tags, val)
	case reflect.Float64:
		return p.parseFloat64(f, tags, val)
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
		return p.parseString(f, tags, val)
	default:
		return fmt.Errorf("unsupported type %s of %s", f.Kind(), f.Type())
	}
}

func (p *parser) parseBool(f reflect.Value, val string) error {
	parsed, ok := map[string]bool{
		"1": true, "0": false,
		"t": true, "f": false,
		"true": true, "false": false,
		"y": true, "n": false,
		"yay": true, "nay": false,
		"yes": true, "no": false,
	}[strings.ToLower(val)]

	if !ok {
		return errors.New("invalid bool value " + val)
	}

	f.Set(reflect.ValueOf(parsed))

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

type float interface{ ~float32 | ~float64 }

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type number interface{ integer | float }

func setSliceOfFloats[F float](val []string, f reflect.Value) error {
	var values []F

	for _, v := range val {
		i, err := parseFloat[F](v)
		if err != nil {
			return err
		}

		values = append(values, i)
	}

	f.Set(reflect.ValueOf(values))

	return nil
}

func setSliceOfIntegers[N integer](val []string, f reflect.Value) error {
	var values []N

	for _, v := range val {
		i, err := parseInteger[N](v)
		if err != nil {
			return err
		}

		values = append(values, i)
	}

	f.Set(reflect.ValueOf(values))

	return nil
}

func lookupNumericTag[N number](cond string, tag reflect.StructTag) (*N, error) {
	value, ok := tag.Lookup(cond)
	if !ok {
		return nil, nil
	}

	m, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing %s value: %w", cond, err)
	}

	m1 := N(m)

	return &m1, nil
}

func parseInteger[N integer](val string) (N, error) {
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return N(i), nil
}

func parseFloat[F float](val string) (F, error) {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, err
	}

	return F(f), nil
}

func setInteger[N integer](f reflect.Value, tag reflect.StructTag, val string) error {
	i, err := parseInteger[N](val)
	if err != nil {
		return err
	}

	if min, err := lookupNumericTag[N]("min", tag); err != nil {
		return fmt.Errorf("parsing min value: %w", err)
	} else if min != nil && i < *min {
		return fmt.Errorf("value %v is less than the minimal value %v", val, *min)
	}

	if max, err := lookupNumericTag[N]("max", tag); err != nil {
		return fmt.Errorf("parsing max value: %w", err)
	} else if max != nil && i < *max {
		return fmt.Errorf("value %v is greater than the minimal value %v", val, *max)
	}

	f.Set(reflect.ValueOf(i))

	return nil
}

func (p *parser) parseFloat32(f reflect.Value, tag reflect.StructTag, val string) error {
	i, err := parseFloat[float32](val)
	if err != nil {
		return err
	}

	if min, err := lookupNumericTag[float32]("min", tag); err != nil {
		return fmt.Errorf("parsing min value: %w", err)
	} else if min != nil && i < *min {
		return fmt.Errorf("value %v is less than the minimal value %v", val, *min)
	}

	if max, err := lookupNumericTag[float32]("max", tag); err != nil {
		return fmt.Errorf("parsing max value: %w", err)
	} else if max != nil && i < *max {
		return fmt.Errorf("value %v is greater than the minimal value %v", val, *max)
	}

	f.Set(reflect.ValueOf(i))

	return nil
}

func (p *parser) parseFloat64(f reflect.Value, tag reflect.StructTag, val string) error {
	i, err := parseFloat[float64](val)
	if err != nil {
		return err
	}

	if min, err := lookupNumericTag[float64]("min", tag); err != nil {
		return fmt.Errorf("parsing min value: %w", err)
	} else if min != nil && i < *min {
		return fmt.Errorf("value %v is less than the minimal value %v", val, *min)
	}

	if max, err := lookupNumericTag[float64]("max", tag); err != nil {
		return fmt.Errorf("parsing max value: %w", err)
	} else if max != nil && i < *max {
		return fmt.Errorf("value %v is greater than the minimal value %v", val, *max)
	}

	f.Set(reflect.ValueOf(i))

	return nil
}

func (p *parser) parseString(f reflect.Value, tag reflect.StructTag, val string) error {
	if pattern, ok := tag.Lookup("pattern"); ok {
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

func (p *parser) parseType(f reflect.Value, tag reflect.StructTag, parser func(string) (reflect.Value, error)) error {
	val, ok := p.tagValue(tag)
	if !ok {
		return nil
	}

	if val == "" {
		return errors.New("value required, but not set")
	}

	v, err := parser(val)
	if err != nil {
		return err
	}

	f.Set(v)

	return nil
}

func (p *parser) tagValue(tag reflect.StructTag) (val string, ok bool) {
	env, ok := tag.Lookup(p.envTag)
	if !ok {
		return "", false // no tag, skip it
	}

	omitEmpty := strings.HasSuffix(env, ",omitempty")
	if omitEmpty {
		env = strings.TrimSuffix(env, ",omitempty")
	}

	val, ok = os.LookupEnv(env)
	if !ok { // value not set
		val, ok = tag.Lookup(p.defaultTag)
		if !ok { // no default value set
			if omitEmpty {
				return "", false // tag is there, but can be empty, skip
			}
			return "", true // tag is there, but empty
		}
	}

	return val, true // tag is there and not empty
}
