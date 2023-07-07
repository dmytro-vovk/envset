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

	if min, err := lookupFloatTag[float32]("min", tag); err != nil {
		return fmt.Errorf("parsing min value: %w", err)
	} else if min != nil && i < *min {
		return fmt.Errorf("value %v is less than the minimal value %v", val, *min)
	}

	if max, err := lookupFloatTag[float32]("max", tag); err != nil {
		return fmt.Errorf("parsing max value: %w", err)
	} else if max != nil && i < *max {
		return fmt.Errorf("value %v is greater than the minimal value %v", val, *max)
	}

	f.Set(reflect.ValueOf(i))

	return nil
}

func (p *parser) setFloat64(f reflect.Value, tag reflect.StructTag, val string) error {
	i, err := parseFloat[float64](val)
	if err != nil {
		return err
	}

	if min, err := lookupFloatTag[float64]("min", tag); err != nil {
		return fmt.Errorf("parsing min value: %w", err)
	} else if min != nil && i < *min {
		return fmt.Errorf("value %v is less than the minimal value %v", val, *min)
	}

	if max, err := lookupFloatTag[float64]("max", tag); err != nil {
		return fmt.Errorf("parsing max value: %w", err)
	} else if max != nil && i < *max {
		return fmt.Errorf("value %v is greater than the minimal value %v", val, *max)
	}

	f.Set(reflect.ValueOf(i))

	return nil
}
