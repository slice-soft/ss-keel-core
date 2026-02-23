package config

import (
	"fmt"
	"os"
	"strconv"
)

func generateError(name string) string {
	return fmt.Sprintf("required environment variable not found: %s", name)
}

// GetEnv gets an environment variable as string. Panics if it doesn't exist.
func GetEnv(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		panic(generateError(name))
	}
	return value
}

// GetEnvOrDefault gets an environment variable or returns the default value.
func GetEnvOrDefault(name, defaultValue string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	return value
}

func GetEnvInt(name string) int {
	value := GetEnv(name)
	result, err := strconv.Atoi(value)
	if err != nil {
		panic(generateError(name))
	}
	return result
}

func GetEnvIntOrDefault(name string, defaultValue int) int {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return result
}

func GetEnvUint(name string) uint {
	value := GetEnv(name)
	result, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		panic(generateError(name))
	}
	return uint(result)
}

func GetEnvBool(name string) bool {
	value := GetEnv(name)
	result, err := strconv.ParseBool(value)
	if err != nil {
		panic(generateError(name))
	}
	return result
}

func GetEnvBoolOrDefault(name string, defaultValue bool) bool {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return result
}
