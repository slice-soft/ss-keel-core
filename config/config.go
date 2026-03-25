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

// GetEnvOrDefault retrieves an environment variable by name and falls back to
// defaultValue when the variable is not set.
func GetEnvOrDefault(name, defaultValue string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
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

// GetEnvIntOrDefault retrieves an integer environment variable and falls back
// to defaultValue when it is not set.
func GetEnvIntOrDefault(name string, defaultValue int) int {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return defaultValue
	}

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

// GetEnvUintOrDefault retrieves an unsigned integer environment variable and
// falls back to defaultValue when it is not set.
func GetEnvUintOrDefault(name string, defaultValue uint) uint {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return defaultValue
	}

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

// GetEnvBoolOrDefault retrieves a boolean environment variable and falls back
// to defaultValue when it is not set.
func GetEnvBoolOrDefault(name string, defaultValue bool) bool {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return defaultValue
	}

	result, err := strconv.ParseBool(value)
	if err != nil {
		panic(generateError(name))
	}
	return result
}

// GetString returns a resolved application setting. It checks exact OS
// environment variables first and then application.properties.
func GetString(key string) string {
	value, ok := lookupSetting(key)
	if !ok {
		panic(generateError(key))
	}
	return value
}

// GetStringOrDefault retrieves a resolved application setting and falls back to
// defaultValue when no value is configured.
func GetStringOrDefault(key, defaultValue string) string {
	value, ok := lookupSetting(key)
	if !ok {
		return defaultValue
	}
	return value
}

// GetInt retrieves a resolved application setting as an integer.
func GetInt(key string) int {
	value := GetString(key)
	result, err := strconv.Atoi(value)
	if err != nil {
		panic(generateError(key))
	}
	return result
}

// GetIntOrDefault retrieves a resolved application setting as an integer and
// falls back to defaultValue when no value is configured.
func GetIntOrDefault(key string, defaultValue int) int {
	value, ok := lookupSetting(key)
	if !ok || value == "" {
		return defaultValue
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		panic(generateError(key))
	}
	return result
}

// GetUint retrieves a resolved application setting as an unsigned integer.
func GetUint(key string) uint {
	value := GetString(key)
	result, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		panic(generateError(key))
	}
	return uint(result)
}

// GetUintOrDefault retrieves a resolved application setting as an unsigned
// integer and falls back to defaultValue when no value is configured.
func GetUintOrDefault(key string, defaultValue uint) uint {
	value, ok := lookupSetting(key)
	if !ok || value == "" {
		return defaultValue
	}

	result, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		panic(generateError(key))
	}
	return uint(result)
}

// GetBool retrieves a resolved application setting as a boolean.
func GetBool(key string) bool {
	value := GetString(key)
	result, err := strconv.ParseBool(value)
	if err != nil {
		panic(generateError(key))
	}
	return result
}

// GetBoolOrDefault retrieves a resolved application setting as a boolean and
// falls back to defaultValue when no value is configured.
func GetBoolOrDefault(key string, defaultValue bool) bool {
	value, ok := lookupSetting(key)
	if !ok || value == "" {
		return defaultValue
	}

	result, err := strconv.ParseBool(value)
	if err != nil {
		panic(generateError(key))
	}
	return result
}
