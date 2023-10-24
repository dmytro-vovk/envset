package envset

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTags(t *testing.T) {
	testCases := []struct {
		tag      reflect.StructTag
		val      string
		exist    bool
		optional bool
	}{
		{
			tag:      "",
			val:      "",
			exist:    false,
			optional: false,
		},
		{
			tag:      `env:"IS_SET"`,
			val:      "is_set",
			exist:    true,
			optional: false,
		},
		{
			tag:      `env:"IS_NOT_SET"`,
			val:      "",
			exist:    true,
			optional: false,
		},
		{
			tag:      `env:"IS_NOT_SET,omitempty"`,
			val:      "",
			exist:    false,
			optional: true,
		},
	}

	t.Setenv("IS_SET", "is_set")

	p := buildParser([]Option{})

	for _, tt := range testCases {
		tt := tt

		t.Run(string(tt.tag), func(t *testing.T) {
			val, exists, optional := p.tagValue(tt.tag)
			assert.Equal(t, tt.exist, exists)
			assert.Equal(t, tt.val, val)
			assert.Equal(t, tt.optional, optional)
		})
	}
}
