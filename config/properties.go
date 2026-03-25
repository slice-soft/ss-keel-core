package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const applicationPropertiesFile = "application.properties"

var (
	propertyPlaceholderPattern = regexp.MustCompile(`\$\{([^{}]+)\}`)

	propertiesMu     sync.RWMutex
	propertiesLoaded bool
	propertiesValues = map[string]string{}
)

// LoadApplicationProperties loads the nearest application.properties file,
// walking up from the current working directory.
func LoadApplicationProperties() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine working directory: %w", err)
	}

	values, err := loadApplicationPropertiesFromDir(dir)
	if err != nil {
		return err
	}

	setApplicationProperties(values)
	return nil
}

func lookupSetting(key string) (string, bool) {
	if value, ok := os.LookupEnv(key); ok {
		return value, true
	}

	ensureApplicationPropertiesLoaded()

	propertiesMu.RLock()
	defer propertiesMu.RUnlock()

	value, ok := propertiesValues[key]
	return value, ok
}

func ensureApplicationPropertiesLoaded() {
	propertiesMu.RLock()
	if propertiesLoaded {
		propertiesMu.RUnlock()
		return
	}
	propertiesMu.RUnlock()

	dir, err := os.Getwd()
	if err != nil {
		setApplicationProperties(map[string]string{})
		return
	}

	values, err := loadApplicationPropertiesFromDir(dir)
	if err != nil {
		setApplicationProperties(map[string]string{})
		return
	}

	setApplicationProperties(values)
}

func setApplicationProperties(values map[string]string) {
	propertiesMu.Lock()
	defer propertiesMu.Unlock()

	propertiesValues = values
	propertiesLoaded = true
}

func loadApplicationPropertiesFromDir(dir string) (map[string]string, error) {
	path, found := findApplicationProperties(dir)
	if !found {
		return nil, fmt.Errorf("%s not found: every Keel project requires an %s at its root", applicationPropertiesFile, applicationPropertiesFile)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	values, err := parseApplicationProperties(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return values, nil
}

func findApplicationProperties(dir string) (string, bool) {
	for {
		candidate := filepath.Join(dir, applicationPropertiesFile)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func parseApplicationProperties(content string) (map[string]string, error) {
	raw := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(content))

	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		separator := strings.IndexAny(line, "=:")
		if separator == -1 {
			return nil, fmt.Errorf("line %d: expected key=value", lineNumber)
		}

		key := strings.TrimSpace(line[:separator])
		value := strings.TrimSpace(line[separator+1:])
		if key == "" {
			return nil, fmt.Errorf("line %d: property key is empty", lineNumber)
		}

		raw[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	resolved := make(map[string]string, len(raw))
	for key := range raw {
		value, err := resolvePropertyValue(key, raw, resolved, map[string]bool{})
		if err != nil {
			return nil, err
		}
		resolved[key] = value
	}

	return resolved, nil
}

func resolvePropertyValue(key string, raw, resolved map[string]string, visiting map[string]bool) (string, error) {
	if value, ok := resolved[key]; ok {
		return value, nil
	}

	if visiting[key] {
		return "", fmt.Errorf("circular property reference detected for %q", key)
	}
	visiting[key] = true
	defer delete(visiting, key)

	value := raw[key]
	var replaceErr error
	resolvedValue := propertyPlaceholderPattern.ReplaceAllStringFunc(value, func(match string) string {
		if replaceErr != nil {
			return ""
		}

		parts := strings.SplitN(strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}"), ":", 2)
		ref := strings.TrimSpace(parts[0])
		fallback := ""
		if len(parts) == 2 {
			fallback = parts[1]
		}

		if envValue, ok := os.LookupEnv(ref); ok {
			return envValue
		}

		if _, ok := raw[ref]; ok {
			refValue, err := resolvePropertyValue(ref, raw, resolved, visiting)
			if err != nil {
				replaceErr = err
				return ""
			}
			return refValue
		}

		return fallback
	})
	if replaceErr != nil {
		return "", replaceErr
	}

	return resolvedValue, nil
}
