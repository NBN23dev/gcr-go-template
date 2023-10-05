package utils

import (
	"fmt"
	"time"
)

// ExecutionTime logs the time that a function took to execute.
func ExecutionTime(start time.Time, name string) {
	elapsed := time.Since(start)

	fmt.Printf("%s took %s", name, elapsed)
}
