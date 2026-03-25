package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// LoadConfig reads application.properties and environment variables to populate
// a typed config struct T. Struct fields must use `keel:"key"` or
// `keel:"key,required"` tags.
//
// Resolution order for each field:
//  1. Exact environment variable named by key
//  2. application.properties entry with the same key
//
// Returns an error listing all missing required variables on startup.
func LoadConfig[T any]() (T, error) {
	ensureApplicationPropertiesLoaded()
	return loadConfigWithLookup[T](lookupSetting)
}

// loadConfigWithLookup is the testable core of LoadConfig.
func loadConfigWithLookup[T any](lookup func(string) (string, bool)) (T, error) {
	var zero T

	var cfg T
	cfgVal := reflect.ValueOf(&cfg).Elem()
	cfgType := cfgVal.Type()

	var missing []string

	for i := range cfgType.NumField() {
		field := cfgType.Field(i)
		fieldVal := cfgVal.Field(i)

		tag := field.Tag.Get("keel")
		if tag == "" || tag == "-" {
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
			return zero, fmt.Errorf("config field %s (%s): %w", field.Name, key, err)
		}
	}

	if len(missing) > 0 {
		return zero, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

// IsDev reports whether the current environment is development.
// Returns true when app.env / APP_ENV is unset, empty, or "development".
func IsDev() bool {
	env := GetStringOrDefault("app.env", GetEnvOrDefault("APP_ENV", "development"))
	return env == "" || env == "development"
}

// IsProd reports whether the current environment is production.
func IsProd() bool {
	return GetStringOrDefault("app.env", GetEnvOrDefault("APP_ENV", "")) == "production"
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
