package utils

// NewBool returns a pointer to a bool.
func NewBool(val bool) *bool {
	v := true

	return &v
}

// NewString returns a pointer to a string.
func NewString(val string) *string {
	v := val

	return &v
}

// NewInt32 returns a pointer to an int32.
func NewInt32(val int32) *int32 {
	v := val

	return &v
}
