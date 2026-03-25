package config

import (
	"fmt"
	"os"
	"strconv"
)

// generateEnvError creates a standard error message for missing environment variables.
func generateEnvError(name string) string {
	return fmt.Sprintf("required environment variable not found: %s", name)
}

// generateConfigError creates a standard error message for missing runtime configuration values.
func generateConfigError(name string) string {
	return fmt.Sprintf("required config value not found: %s", name)
}

// GetEnv retrieves an environment variable by name and returns its string value.
// It panics if the environment variable is not set.
func GetEnv(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		panic(generateEnvError(name))
	}
	return value
}

// GetEnvInt retrieves an environment variable by name and returns its integer value.
// It panics if the environment variable is not set or cannot be parsed as an integer.
func GetEnvInt(name string) int {
	value := GetEnv(name)
	result, err := strconv.Atoi(value)
	if err != nil {
		panic(generateEnvError(name))
	}
	return result
}

// GetEnvUint retrieves an environment variable by name and returns its unsigned integer value.
// It panics if the environment variable is not set or cannot be parsed as an unsigned integer.
func GetEnvUint(name string) uint {
	value := GetEnv(name)
	result, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		panic(generateEnvError(name))
	}
	return uint(result)
}

// GetEnvBool retrieves an environment variable by name and returns its boolean value.
// It panics if the environment variable is not set or cannot be parsed as a boolean.
func GetEnvBool(name string) bool {
	value := GetEnv(name)
	result, err := strconv.ParseBool(value)
	if err != nil {
		panic(generateEnvError(name))
	}
	return result
}

// GetString returns a resolved application setting. It checks exact OS
// environment variables first and then application.properties.
func GetString(key string) string {
	value, ok := lookupSetting(key)
	if !ok {
		panic(generateConfigError(key))
	}
	return value
}

// GetInt retrieves a resolved application setting as an integer.
func GetInt(key string) int {
	value := GetString(key)
	result, err := strconv.Atoi(value)
	if err != nil {
		panic(generateConfigError(key))
	}
	return result
}

// GetUint retrieves a resolved application setting as an unsigned integer.
func GetUint(key string) uint {
	value := GetString(key)
	result, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		panic(generateConfigError(key))
	}
	return uint(result)
}

// GetBool retrieves a resolved application setting as a boolean.
func GetBool(key string) bool {
	value := GetString(key)
	result, err := strconv.ParseBool(value)
	if err != nil {
		panic(generateConfigError(key))
	}
	return result
}

// LookupString returns a resolved application setting when present.
// It checks exact OS environment variables first and then application.properties.
func LookupString(key string) (string, bool) {
	return lookupSetting(key)
}

// LookupInt returns a resolved application setting as an integer when present.
func LookupInt(key string) (int, bool) {
	value, ok := lookupSetting(key)
	if !ok || value == "" {
		return 0, false
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		panic(generateConfigError(key))
	}
	return result, true
}

// LookupUint returns a resolved application setting as an unsigned integer when present.
func LookupUint(key string) (uint, bool) {
	value, ok := lookupSetting(key)
	if !ok || value == "" {
		return 0, false
	}

	result, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		panic(generateConfigError(key))
	}
	return uint(result), true
}

// LookupBool returns a resolved application setting as a boolean when present.
func LookupBool(key string) (bool, bool) {
	value, ok := lookupSetting(key)
	if !ok || value == "" {
		return false, false
	}

	result, err := strconv.ParseBool(value)
	if err != nil {
		panic(generateConfigError(key))
	}
	return result, true
}
