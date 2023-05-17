package envset_test

import (
	"os"
	"testing"
	"time"

	"github.com/dmytro-vovk/envset"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	type MyStruct struct {
		String         string `env:"STRING"`
		NoTag          string
		StringPointer  *string `env:"STRING_POINTER"`
		StringPointer2 *string `env:"STRING_POINTER_UNSET,omitempty"`
		Struct         struct {
			String string `env:"INTERNAL_STRING"`
		}
		StructPtr *struct {
			String string `env:"INTERNAL_STRING_2"`
		}
		Time     time.Time     `env:"TIME"`
		Duration time.Duration `env:"DURATION"`
	}

	require.NoError(t, os.Setenv("STRING", "a string"))
	require.NoError(t, os.Setenv("STRING_POINTER", "string pointer"))
	require.NoError(t, os.Setenv("INTERNAL_STRING", "nested string"))
	require.NoError(t, os.Setenv("INTERNAL_STRING_2", "nested string 2"))
	require.NoError(t, os.Setenv("DURATION", "2s"))
	now := time.Now().In(time.UTC)
	require.NoError(t, os.Setenv("TIME", now.String()))

	var ms MyStruct

	require.Error(t, envset.ErrStructExpected, envset.Set(ms))
	require.Error(t, envset.ErrStructExpected, envset.Set(false))

	var i int
	require.Error(t, envset.ErrStructExpected, envset.Set(&i))

	require.NoError(t, envset.Set(
		&ms,
		envset.WithTypeParser(time.ParseDuration),
		envset.WithTypeParser(func(val string) (time.Time, error) {
			return time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", val)
		}),
	))

	assert.Equal(t, "a string", ms.String)
	assert.Equal(t, "string pointer", *ms.StringPointer)
	assert.Nil(t, ms.StringPointer2)
	assert.Equal(t, "nested string", ms.Struct.String)
	assert.Equal(t, 2*time.Second, ms.Duration)
	assert.Equal(t, now, ms.Time)

	t.Logf("%# v", pretty.Formatter(ms))
}
