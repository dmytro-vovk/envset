package envset_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/dmytro-vovk/envset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	type MyStruct struct {
		String   string `env:"STRING"`
		NoTag    string
		NoEnvTag string `json:"no_env_tag"`

		StringPointer  *string `env:"STRING_POINTER"`
		StringPointer2 *string `env:"STRING_POINTER_UNSET,omitempty"`

		BY1 bool `env:"B" default:"true"`
		BY2 bool `env:"B" default:"y"`
		BN  bool `env:"B" default:"no"`

		Struct struct {
			String string `env:"INTERNAL_STRING"`
		}

		StructPtr *struct {
			String string `env:"INTERNAL_STRING_2"`
		}

		Time     time.Time     `env:"TIME"`
		Duration time.Duration `env:"DURATION"`

		private string `env:"STRING"`

		F   float64 `env:"F"`
		F32 float32 `env:"F"`

		I   int   `env:"I" min:"-1"`
		I8  int8  `env:"I8"`
		I16 int16 `env:"I16"`
		I32 int32 `env:"I32"`
		I64 int64 `env:"I64"`

		U   uint   `env:"U"`
		U8  uint8  `env:"U8"`
		U16 uint16 `env:"U16"`
		U32 uint32 `env:"U32"`
		U64 uint64 `env:"U64"`

		FList []float32 `env:"F_LIST"`
		IList []int     `env:"I_LIST"`
	}

	require.NoError(t, os.Setenv("STRING", "a string"))
	require.NoError(t, os.Setenv("STRING_POINTER", "string pointer"))
	require.NoError(t, os.Setenv("INTERNAL_STRING", "nested string"))
	require.NoError(t, os.Setenv("INTERNAL_STRING_2", "nested string 2"))
	require.NoError(t, os.Setenv("DURATION", (2*time.Second).String()))
	require.NoError(t, os.Setenv("F", "1.2345"))
	require.NoError(t, os.Setenv("F_LIST", "1.2,2.3,3.4"))
	require.NoError(t, os.Setenv("I_LIST", "1,2,3"))

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

	require.NotPanics(t, func() { _ = envset.Set(&struct{}{}) })

	var i int
	require.Panics(t, func() { _ = envset.Set(&i) })

	var ms MyStruct

	require.NoError(t, envset.Set(
		&ms,
		envset.WithTypeParser(time.ParseDuration),
		envset.WithTypeParser(func(val string) (time.Time, error) {
			return time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", val)
		}),
	))

	expected := MyStruct{
		String:         "a string",
		NoTag:          "",
		NoEnvTag:       "",
		StringPointer:  ptr("string pointer"),
		StringPointer2: nil,
		BY1:            true,
		BY2:            true,
		BN:             false,
		Struct: struct {
			String string `env:"INTERNAL_STRING"`
		}{
			String: "nested string",
		},
		StructPtr: &struct {
			String string `env:"INTERNAL_STRING_2"`
		}{
			String: "nested string 2",
		},
		Time:     now,
		Duration: 2 * time.Second,
		private:  "",
		F:        1.2345,
		F32:      1.2345,
		I:        -1,
		I8:       -8,
		I16:      -16,
		I32:      -32,
		I64:      -63,
		U:        1,
		U8:       8,
		U16:      16,
		U32:      32,
		U64:      64,
		FList:    []float32{1.2, 2.3, 3.4},
		IList:    []int{1, 2, 3},
	}

	assert.Equal(t, expected, ms)
}

func ptr[T any](v T) *T { return &v }

func TestNoTag(t *testing.T) {
	type T struct {
		A int
		B string
		C float64
		D bool
		E struct {
			F []int
			G []string
		}
	}

	var v T
	require.NoError(t, envset.Set(&v))
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
	require.ErrorIs(t, envset.NewMissingValueError("a"), envset.Set(&v))
}

func TestOmitEmptyWithDefault(t *testing.T) {
	type T struct {
		A int `env:"a,omitempty" default:"10"`
	}

	var v T
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, 10, v.A)
}

func TestOmitEmptyWithNoDefault(t *testing.T) {
	type T struct {
		A int `env:"a,omitempty"`
	}

	var v T
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, 0, v.A)
}

func TestOmitEmptyWithDefault2(t *testing.T) {
	type T struct {
		A int `env:"a" default:"10"`
	}

	var v T
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, 10, v.A)
}

func TestOmitEmptyWithDefault3(t *testing.T) {
	type T struct {
		A int `e:"a" d:"10"`
	}

	_ = os.Setenv("a", "15")

	var v T
	require.NoError(t, envset.Set(&v, envset.WithEnvTag("e"), envset.WithDefaultTag("d")))
	assert.Equal(t, 15, v.A)
}

func TestUnsupportedType(t *testing.T) {
	type T struct {
		C chan bool `env:"C" default:"1"`
	}

	var v T

	require.Error(t, envset.Set(&v))
}

func TestBool(t *testing.T) {
	type T struct {
		B bool `env:"b"`
	}

	_ = os.Setenv("b", "15")

	var v T
	require.Error(t, envset.Set(&v))
}

func TestBool2(t *testing.T) {
	type T struct {
		B bool `env:"b" default:"xyz"`
	}

	var v T

	require.Error(t, envset.Set(&v))
}

func TestCustomBool(t *testing.T) {
	type T struct {
		BT bool `env:"bt"`
		BF bool `env:"bf"`
	}

	_ = os.Setenv("bt", "так")
	_ = os.Setenv("bf", "ні")

	var v T

	require.NoError(t, envset.Set(&v, envset.WithCustomBools("так", "ні")))

	assert.True(t, v.BT)
	assert.False(t, v.BF)
}

func TestString(t *testing.T) {
	type T struct {
		S string `env:"s" pattern:"[a-" default:"abc"`
	}

	var v T

	require.ErrorContains(t, envset.Set(&v), "error parsing regexp")
}

func TestString2(t *testing.T) {
	type T struct {
		S string `env:"s" pattern:"^\\d$"`
	}

	_ = os.Setenv("s", "15abc")

	var v T

	require.ErrorIs(t, envset.ErrInvalidValue, envset.Set(&v))
}

func TestSliceSeparator(t *testing.T) {
	type T struct {
		S []string `env:"s"`
	}

	_ = os.Setenv("s", "a•b•c•d")

	var v T

	require.NoError(t, envset.Set(&v, envset.WithSliceSeparator("•")))

	assert.Equal(t, []string{"a", "b", "c", "d"}, v.S)
}

func TestCustomType(t *testing.T) {
	type C string
	type T struct {
		C C `env:"c"`
	}

	t.Setenv("c", "bar")

	var v T
	require.NoError(t, envset.Set(&v, envset.WithTypeParser(func(string) (C, error) {
		return "foo", nil
	})))

	assert.Equal(t, C("foo"), v.C)
}

func TestCustomTypeNoTag(t *testing.T) {
	type C string
	type T struct {
		C C
	}

	var v T
	require.NoError(t, envset.Set(&v))
}

func TestCustomTypeParseError(t *testing.T) {
	type C string
	type T struct {
		C C `env:"c"`
	}

	t.Setenv("c", "bar")

	var v T
	require.ErrorContains(t, envset.Set(&v, envset.WithTypeParser(func(string) (C, error) {
		return "", errors.New("boo")
	})), "boo")
}

func TestCustomOptionalTypeParseError(t *testing.T) {
	type C string
	type T struct {
		C C `env:"c,omitempty"`
	}
	var v T
	require.NoError(t, envset.Set(&v, envset.WithTypeParser(func(string) (C, error) {
		return "", errors.New("boo")
	})))
}

func TestCustomTypeString(t *testing.T) {
	type C string
	type A = string
	type T struct {
		C C `env:"c" default:"foo"`
		A A `env:"c,omitempty" default:"bar"`
	}
	var v T
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, C("foo"), v.C)
	assert.Equal(t, "bar", v.A)
}

func TestCustomTypeInt(t *testing.T) {
	type C int
	type A = int
	type T struct {
		C C `env:"c" default:"100"`
		A A `env:"c" default:"200"`
	}
	var v T
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, C(100), v.C)
	assert.Equal(t, 200, v.A)
}

func TestCustomTypeByte(t *testing.T) {
	type C byte
	type A = byte
	type T struct {
		C C `env:"c" default:"100"`
		A A `env:"c" default:"200"`
	}
	var v T
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, C(100), v.C)
	assert.Equal(t, byte(200), v.A)
}

func TestCustomTypeFloat(t *testing.T) {
	type C float64
	type A = float32
	type T struct {
		C C `env:"c" default:"10.0"`
		A A `env:"c" default:"20.0"`
	}
	var v T
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, C(10.0), v.C)
	assert.Equal(t, float32(20.0), v.A)
}

func TestCustomTypeBool(t *testing.T) {
	type C bool
	type A = bool
	type T struct {
		C C `env:"c" default:"y"`
		A A `env:"c" default:"n"`
	}
	var v T
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, C(true), v.C)
	assert.Equal(t, false, v.A)
}

func TestEmptyEnvValue(t *testing.T) {
	t.Setenv("a", "")

	type T struct {
		A string `env:"a"`
	}

	var v T
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, "", v.A)
}

func TestSkipDefault(t *testing.T) {
	type T struct {
		A string `env:"a" default:"three"`
	}

	var v T
	v.A = "one"
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, "one", v.A)
}

func TestSkipEnv(t *testing.T) {
	t.Setenv("a", "two")

	type T struct {
		A string `env:"a"`
	}

	var v T
	v.A = "one"
	require.NoError(t, envset.Set(&v))
	assert.Equal(t, "one", v.A)
}
