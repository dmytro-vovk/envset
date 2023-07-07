package envset

import "errors"

var (
	ErrInvalidValue      = errors.New("invalid value")
	ErrMissingValue      = errors.New("value required, but not set")
	ErrStructPtrExpected = errors.New("pointer to struct expected")
)
