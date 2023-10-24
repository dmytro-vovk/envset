package envset

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTags(t *testing.T) {
	testCases := []struct {
		tag      reflect.StructTag
		key      string
		exist    bool
		optional bool
	}{
		{
			tag:      "",
			key:      "",
			exist:    false,
			optional: false,
		},
		{
			tag:      `env:"IS_SET"`,
			key:      "IS_SET",
			exist:    true,
			optional: false,
		},
		{
			tag:      `env:"IS_NOT_SET"`,
			key:      "IS_NOT_SET",
			exist:    true,
			optional: false,
		},
		{
			tag:      `env:"IS_NOT_SET,omitempty"`,
			key:      "IS_NOT_SET",
			exist:    true,
			optional: true,
		},
	}

	t.Setenv("IS_SET", "is_set")

	p := buildParser([]Option{})

	for _, tt := range testCases {
		tt := tt

		t.Run(string(tt.tag), func(t *testing.T) {
			key, exists, optional := p.tagKey(tt.tag)
			assert.Equal(t, tt.exist, exists)
			assert.Equal(t, tt.key, key)
			assert.Equal(t, tt.optional, optional)
		})
	}
}
