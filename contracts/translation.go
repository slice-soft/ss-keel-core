package contracts

// Translator is the contract for i18n providers (e.g. ss-keel-i18n).
type Translator interface {
	T(locale, key string, args ...any) string
	Locales() []string
}
