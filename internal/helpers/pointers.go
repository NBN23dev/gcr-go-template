package helpers

// GetPointer returns a pointer to the value passed as argument.
func GetPointer[T any](value T) *T {
	v := value

	return &v
}
