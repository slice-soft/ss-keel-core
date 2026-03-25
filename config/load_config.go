package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// LoadConfig reads application.properties and environment variables to populate
// a typed config struct T. Struct fields must use `keel:"key"` or
// `keel:"key,required"` tags.
//
// Resolution order for each field:
//  1. Process environment, including values loaded automatically from the
//     nearest .env file when present
//  2. application.properties entry with the same key
//
// Nested structs are traversed recursively, so runtime config can mirror the
// application's real shape instead of being flattened into key-by-key reads.
//
// Returns an error listing all missing required config values on startup.
func LoadConfig[T any]() (T, error) {
	ensureApplicationPropertiesLoaded()
	return loadConfigWithLookup[T](lookupSetting)
}

// MustLoadConfig loads a typed runtime config or panics during startup.
func MustLoadConfig[T any]() T {
	cfg, err := LoadConfig[T]()
	if err != nil {
		panic(fmt.Sprintf("failed to load runtime config: %v", err))
	}
	return cfg
}

// loadConfigWithLookup is the testable core of LoadConfig.
func loadConfigWithLookup[T any](lookup func(string) (string, bool)) (T, error) {
	var zero T

	var cfg T
	missing, err := loadStructWithLookup(reflect.ValueOf(&cfg).Elem(), lookup, "")
	if err != nil {
		return zero, err
	}

	if len(missing) > 0 {
		return zero, fmt.Errorf("missing required config values: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func loadStructWithLookup(v reflect.Value, lookup func(string) (string, bool), path string) ([]string, error) {
	t := v.Type()
	var missing []string

	for i := range t.NumField() {
		field := t.Field(i)
		fieldVal := v.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		fieldPath := field.Name
		if path != "" {
			fieldPath = path + "." + field.Name
		}

		tag := field.Tag.Get("keel")
		if tag == "-" {
			continue
		}

		if tag == "" {
			if fieldVal.Kind() == reflect.Struct {
				nestedMissing, err := loadStructWithLookup(fieldVal, lookup, fieldPath)
				if err != nil {
					return nil, err
				}
				missing = append(missing, nestedMissing...)
			}
			continue
		}

		parts := strings.SplitN(tag, ",", 2)
		key := parts[0]
		required := len(parts) > 1 && parts[1] == "required"

		raw, ok := lookup(key)
		if !ok {
			if required {
				missing = append(missing, key)
			}
			continue
		}

		if err := setField(fieldVal, raw); err != nil {
			return nil, fmt.Errorf("config field %s (%s): %w", fieldPath, key, err)
		}
	}

	return missing, nil
}

// IsDev reports whether the current environment is development.
// Returns true when app.env / APP_ENV is unset, empty, or "development".
func IsDev() bool {
	env, ok := lookupSetting("app.env")
	if !ok {
		env, ok = os.LookupEnv("APP_ENV")
	}
	if !ok {
		return true
	}
	return env == "" || env == "development"
}

// IsProd reports whether the current environment is production.
func IsProd() bool {
	env, ok := lookupSetting("app.env")
	if !ok {
		env, ok = os.LookupEnv("APP_ENV")
	}
	return ok && env == "production"
}

// setField converts the raw string s into the appropriate Go type and assigns it to v.
func setField(v reflect.Value, s string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("expected integer, got %q", s)
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fmt.Errorf("expected unsigned integer, got %q", s)
		}
		v.SetUint(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("expected boolean, got %q", s)
		}
		v.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("expected float, got %q", s)
		}
		v.SetFloat(f)
	default:
		return fmt.Errorf("unsupported field type %s", v.Kind())
	}
	return nil
}
