package envset

import (
	"fmt"
	"reflect"
	"strconv"
)

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
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

func lookupIntegerTag[N integer](tagName string, tag reflect.StructTag) (*N, error) {
	value, ok := tag.Lookup(tagName)
	if !ok {
		return nil, nil
	}

	m, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("parsing %s value: %w", tagName, err)
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

func setInteger[N integer](f reflect.Value, tag reflect.StructTag, val string) error {
	i, err := parseInteger[N](val)
	if err != nil {
		return err
	}

	if min, err := lookupIntegerTag[N]("min", tag); err != nil {
		return fmt.Errorf("parsing min value: %w", err)
	} else if min != nil && i < *min {
		return fmt.Errorf("value %v is less than the minimal value %v", val, *min)
	}

	if max, err := lookupIntegerTag[N]("max", tag); err != nil {
		return fmt.Errorf("parsing max value: %w", err)
	} else if max != nil && i < *max {
		return fmt.Errorf("value %v is greater than the minimal value %v", val, *max)
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
