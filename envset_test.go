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
		private  string        `env:"STRING"`
		F        float64       `env:"F"`
		F32      float32       `env:"F"`
		I        int           `env:"I"`
		I8       int8          `env:"I8"`
		I16      int16         `env:"I16"`
		I32      int32         `env:"I32"`
		I64      int64         `env:"I64"`
		U        uint          `env:"U"`
		U8       uint8         `env:"U8"`
		U16      uint16        `env:"U16"`
		U32      uint32        `env:"U32"`
		U64      uint64        `env:"U64"`
	}

	require.NoError(t, os.Setenv("STRING", "a string"))
	require.NoError(t, os.Setenv("STRING_POINTER", "string pointer"))
	require.NoError(t, os.Setenv("INTERNAL_STRING", "nested string"))
	require.NoError(t, os.Setenv("INTERNAL_STRING_2", "nested string 2"))
	require.NoError(t, os.Setenv("DURATION", (2*time.Second).String()))
	require.NoError(t, os.Setenv("F", "1.2345"))

	require.NoError(t, os.Setenv("I", "-1"))
	require.NoError(t, os.Setenv("I8", "-8"))
	require.NoError(t, os.Setenv("I16", "-16"))
	require.NoError(t, os.Setenv("I32", "-32"))
	require.NoError(t, os.Setenv("I64", "-63"))

	require.NoError(t, os.Setenv("U", "1"))
	require.NoError(t, os.Setenv("U8", "8"))
	require.NoError(t, os.Setenv("U16", "16"))
	require.NoError(t, os.Setenv("U32", "32"))
	require.NoError(t, os.Setenv("U64", "64"))

	now := time.Now().In(time.UTC)
	require.NoError(t, os.Setenv("TIME", now.String()))

	var ms MyStruct

	require.Panics(t, func() { envset.Set(ms) })
	require.Panics(t, func() { envset.Set(false) })
	require.Panics(t, func() { envset.Set(nil) })
	require.NotPanics(t, func() { envset.Set(&struct{}{}) })

	var i int
	require.Panics(t, func() { envset.Set(&i) })

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
	assert.Equal(t, 1.2345, ms.F)

	t.Logf("%# v", pretty.Formatter(ms))
}

func TestOmitEmpty(t *testing.T) {
	type T struct {
		A int `env:"a,omitempty"`
	}

	var v T
	require.NoError(t, envset.Set(&v))
}

func TestOmitEmpty2(t *testing.T) {
	type T struct {
		A int `env:"a"`
	}

	var v T
	require.Error(t, envset.Set(&v))
}

func TestOmitEmptyWithDefault(t *testing.T) {
	type T struct {
		A int `env:"a,omitempty" default:"10"`
	}

	var v T
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, 10, v.A)
}

func TestOmitEmptyWithDefault2(t *testing.T) {
	type T struct {
		A int `env:"a" default:"10"`
	}

	var v T
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, 10, v.A)
}
