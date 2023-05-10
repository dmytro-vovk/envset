package envset

type Option func(*parser)

func WithSliceSeparator(sep string) Option {
	return func(p *parser) {
		p.sliceSeparator = sep
	}
}

func WithEnvTag(tag string) Option {
	return func(p *parser) {
		p.envTag = tag
	}
}

func WithDefaultTag(tag string) Option {
	return func(p *parser) {
		p.defaultTag = tag
	}
}
