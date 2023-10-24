package envset

import (
	"errors"
)

var (
	ErrInvalidValue      = errors.New("invalid value")
	ErrStructPtrExpected = errors.New("pointer to struct expected")
)

type MissingValueError struct {
	value string
}

func NewMissingValueError(value string) MissingValueError {
	return MissingValueError{value: value}
}

func (err MissingValueError) Error() string {
	return "value required, but not set: " + err.value
}
