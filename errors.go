package envset

import "errors"

var (
	ErrInvalidValue   = errors.New("invalid value")
	ErrStructExpected = errors.New("pointer to struct expected")
)
