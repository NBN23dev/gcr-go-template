package helpers

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// ErrNotFound represents an error when the env var is not found.
type ErrNotFound struct {
	key string
}

func (err ErrNotFound) Error() string {
	return fmt.Sprintf("utils: env var %s not found", err.key)
}

// ErrInvalidType represents an error when the type of the env var is invalid.
type ErrInvalidType struct {
	typ string
}

func (err ErrInvalidType) Error() string {
	return fmt.Sprintf("utils: invalid type %s", err.typ)
}

// Value represents a type that can be an int or a string
type Value interface{ string | int }

// GetEnv returns the value associated to an env var.
// If the env var not exists returns ErrNotFound.
func GetEnv[V Value](key string) (V, error) {
	val, exists := os.LookupEnv(key)

	var value V

	switch any(value).(type) {
	case string:
		if !exists {
			return any(val).(V), ErrNotFound{key}
		}

		return any(val).(V), nil
	case int:
		if !exists {
			return any(0).(V), ErrNotFound{key}
		}

		iVal, err := strconv.Atoi(val)

		return any(iVal).(V), err
	default:
		var nilValue V

		return nilValue, ErrInvalidType{fmt.Sprintf("%T", value)}
	}

}

// GetEnv returns the value associated to an env var.
// If the env var not exists returns the provided default value.
func GetEnvOr[V Value](key string, def V) (V, error) {
	val, err := GetEnv[V](key)

	if errors.Is(err, ErrNotFound{key}) {
		return def, nil
	}

	return val, err
}
