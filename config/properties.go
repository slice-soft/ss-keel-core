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
const dotEnvFile = ".env"

var (
	propertyPlaceholderPattern = regexp.MustCompile(`\$\{([^{}]+)\}`)

	propertiesMu     sync.RWMutex
	propertiesLoaded bool
	propertiesValues = map[string]string{}

	dotEnvMu     sync.RWMutex
	dotEnvLoaded bool
)

// LoadApplicationProperties loads the nearest application.properties file,
// walking up from the current working directory.
func LoadApplicationProperties() error {
	ensureDotEnvLoaded()

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
	ensureDotEnvLoaded()

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
	ensureDotEnvLoaded()

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

func ensureDotEnvLoaded() {
	dotEnvMu.RLock()
	if dotEnvLoaded {
		dotEnvMu.RUnlock()
		return
	}
	dotEnvMu.RUnlock()

	dir, err := os.Getwd()
	if err != nil {
		setDotEnvLoaded()
		return
	}

	_ = loadDotEnvFromDir(dir)
	setDotEnvLoaded()
}

func setApplicationProperties(values map[string]string) {
	propertiesMu.Lock()
	defer propertiesMu.Unlock()

	propertiesValues = values
	propertiesLoaded = true
}

func setDotEnvLoaded() {
	dotEnvMu.Lock()
	defer dotEnvMu.Unlock()

	dotEnvLoaded = true
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

func loadDotEnvFromDir(dir string) error {
	path, found := findNearestFile(dir, dotEnvFile)
	if !found {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	for key, value := range parseDotEnv(string(data)) {
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set %s from %s: %w", key, path, err)
		}
	}

	return nil
}

func findApplicationProperties(dir string) (string, bool) {
	return findNearestFile(dir, applicationPropertiesFile)
}

func findNearestFile(dir, name string) (string, bool) {
	for {
		candidate := filepath.Join(dir, name)
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

func parseDotEnv(content string) map[string]string {
	values := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimPrefix(line, "export ")
		separator := strings.Index(line, "=")
		if separator <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:separator])
		if key == "" {
			continue
		}

		value := strings.TrimSpace(line[separator+1:])
		values[key] = stripWrappingQuotes(value)
	}

	return values
}

func stripWrappingQuotes(value string) string {
	if len(value) < 2 {
		return value
	}

	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return value[1 : len(value)-1]
	}

	return value
}
