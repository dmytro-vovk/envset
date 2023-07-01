package envset

import "errors"

var (
	ErrInvalidValue      = errors.New("invalid value")
	ErrStructPtrExpected = errors.New("pointer to struct expected")
)
