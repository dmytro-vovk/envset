package envset

import "reflect"

type Option func(*parser)

// WithSliceSeparator sets a separator for slice values.
func WithSliceSeparator(sep string) Option {
	return func(p *parser) {
		p.sliceSeparator = sep
	}
}

// WithEnvTag sets tag that will be used to get env var name.
func WithEnvTag(tag string) Option {
	return func(p *parser) {
		p.envTag = tag
	}
}

// WithDefaultTag sets tag that will be used to lookup default value.
func WithDefaultTag(tag string) Option {
	return func(p *parser) {
		p.defaultTag = tag
	}
}

// WithTypeParser adds a parser for a custom type.
func WithTypeParser[T any](fn func(val string) (T, error)) Option {
	return func(p *parser) {
		var t T
		vt := reflect.TypeOf(t)
		p.customTypes[vt.String()] = func(val string) (reflect.Value, error) {
			v, err := fn(val)
			return reflect.ValueOf(v), err
		}
	}
}
