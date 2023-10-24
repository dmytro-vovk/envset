package envset

import (
	"fmt"
	"reflect"
	"strconv"
)

type float interface{ ~float32 | ~float64 }

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

func lookupFloatTag[N float](tagName string, tag reflect.StructTag) (*N, error) {
	value, ok := tag.Lookup(tagName)
	if !ok {
		return nil, nil
	}

	m, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing %s value: %w", tagName, err)
	}

	m1 := N(m)

	return &m1, nil
}

func parseFloat[F float](val string) (F, error) {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, err
	}

	return F(f), nil
}

func (p *parser) setFloat32(f reflect.Value, tag reflect.StructTag, val string) error {
	i, err := parseFloat[float32](val)
	if err != nil {
		return err
	}

	if minValue, err := lookupFloatTag[float32]("min", tag); err != nil {
		return fmt.Errorf("parsing min value: %w", err)
	} else if minValue != nil && i < *minValue {
		return fmt.Errorf("value %v is less than the minimal value %v", val, *minValue)
	}

	if maxValue, err := lookupFloatTag[float32]("max", tag); err != nil {
		return fmt.Errorf("parsing max value: %w", err)
	} else if maxValue != nil && i < *maxValue {
		return fmt.Errorf("value %v is greater than the minimal value %v", val, *maxValue)
	}

	f.Set(reflect.ValueOf(i))

	return nil
}

func (p *parser) setFloat64(f reflect.Value, tag reflect.StructTag, val string) error {
	i, err := parseFloat[float64](val)
	if err != nil {
		return err
	}

	if minValue, err := lookupFloatTag[float64]("min", tag); err != nil {
		return fmt.Errorf("parsing min value: %w", err)
	} else if minValue != nil && i < *minValue {
		return fmt.Errorf("value %v is less than the minimal value %v", val, *minValue)
	}

	if maxValue, err := lookupFloatTag[float64]("max", tag); err != nil {
		return fmt.Errorf("parsing max value: %w", err)
	} else if maxValue != nil && i < *maxValue {
		return fmt.Errorf("value %v is greater than the minimal value %v", val, *maxValue)
	}

	v := reflect.ValueOf(i)
	if v.Type() == f.Type() {
		f.Set(v)
	} else if v.CanConvert(f.Type()) {
		f.Set(v.Convert(f.Type()))
	} else {
		return fmt.Errorf("value of type %s is not assignable to field of type %s", v.Type(), f.Type())
	}

	return nil
}
