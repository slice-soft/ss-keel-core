package core

// Translator is the contract for i18n providers (e.g. ss-keel-i18n).
type Translator interface {
	T(locale, key string, args ...any) string
	Locales() []string
}

// Lang extracts the language from the Accept-Language header.
// Returns "en" if the header is absent or empty.
func (c *Ctx) Lang() string {
	lang := c.Get("Accept-Language")
	if lang == "" {
		return "en"
	}
	// Use only the primary tag (e.g. "en-US,en;q=0.9" → "en-US").
	for i := 0; i < len(lang); i++ {
		if lang[i] == ',' || lang[i] == ';' {
			return lang[:i]
		}
	}
	return lang
}

// T translates a key using the configured Translator.
// Returns the key unchanged if no Translator is registered.
func (c *Ctx) T(key string, args ...any) string {
	t, ok := c.Locals("_keel_translator").(Translator)
	if !ok || t == nil {
		return key
	}
	return t.T(c.Lang(), key, args...)
}
