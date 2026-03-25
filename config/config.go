package config

import (
	"fmt"
	"os"
	"strconv"
)

// generateError creates a standard error message for missing environment variables.
func generateError(name string) string {
	return fmt.Sprintf("required environment variable not found: %s", name)
}

// GetEnv retrieves an environment variable by name and returns its string value.
// It panics if the environment variable is not set.
func GetEnv(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		panic(generateError(name))
	}
	return value
}

// GetEnvInt retrieves an environment variable by name and returns its integer value.
// It panics if the environment variable is not set or cannot be parsed as an integer.
func GetEnvInt(name string) int {
	value := GetEnv(name)
	result, err := strconv.Atoi(value)
	if err != nil {
		panic(generateError(name))
	}
	return result
}

// GetEnvUint retrieves an environment variable by name and returns its unsigned integer value.
// It panics if the environment variable is not set or cannot be parsed as an unsigned integer.
func GetEnvUint(name string) uint {
	value := GetEnv(name)
	result, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		panic(generateError(name))
	}
	return uint(result)
}

// GetEnvBool retrieves an environment variable by name and returns its boolean value.
// It panics if the environment variable is not set or cannot be parsed as a boolean.
func GetEnvBool(name string) bool {
	value := GetEnv(name)
	result, err := strconv.ParseBool(value)
	if err != nil {
		panic(generateError(name))
	}
	return result
}
