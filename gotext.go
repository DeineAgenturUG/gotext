/*
Package gotext implements GNU gettext utilities.

For quick/simple translations you can use the package level functions directly.

    import (
	    "fmt"
	    "github.com/leonelquinteros/gotext"
    )

    func main() {
        // Configure package
        gotext.Configure("/path/to/locales/root/dir", "en_UK", "domain-name")

        // Translate text from default domain
        fmt.Println(gotext.Get("My text on 'domain-name' domain"))

        // Translate text from a different domain without reconfigure
        fmt.Println(gotext.GetD("domain2", "Another text on a different domain"))
    }

*/
package gotext

import (
	"fmt"
	"sync"

	"github.com/leonelquinteros/gotext/format"
)

// Sprintf alias from submodule
var Sprintf = format.Sprintf

// Global environment variables
type config struct {
	sync.RWMutex

	// Default domain to look at when no domain is specified. Used by package level functions.
	loadDomains []string
	domain      string

	// Language set.
	loadLanguages []string
	language      string

	// Path to library directory where all locale directories and translation files are.
	library string

	// Storage for package level methods
	storage map[string]*Locale
}

var (
	globalConfig *config
	once         sync.Once

	DefaultDomain = "default"
)

// Create Instance default configuration
func GetInstance(loadDomains, loadLanguages []string, defaultDomain, defaultLanguage, library string) *config {
	once.Do(func() {
		globalConfig = &config{
			loadDomains:   loadDomains,
			domain:        defaultDomain,
			loadLanguages: loadLanguages,
			language:      defaultLanguage,
			library:       library,
		}
		globalConfig.loadStorage(true)
	})
	return globalConfig
}

// loadStorage creates a new Locale object at package level based on the Global variables settings.
// It's called automatically when trying to use Get or GetD methods.
func (c *config) loadStorage(force bool) *config {
	c.Lock()

	if c.storage == nil {
		c.storage = make(map[string]*Locale)
	}
	if c.storage[c.language] == nil || force {
		c.storage[c.language] = NewLocale(c.library, c.language)
	}
	if _, ok := c.storage[c.language].domains[c.domain]; !ok || force {
		c.storage[c.language].AddDomain(c.domain)
	}
	for _, language := range c.loadLanguages {
		if c.storage[language] == nil || force {
			c.storage[language] = NewLocale(c.library, c.language)
		}
		for _, domain := range c.loadDomains {
			if _, ok := c.storage[language].domains[domain]; !ok || force {
				c.storage[language].AddDomain(domain)
			}
		}
	}

	c.Unlock()
	return c
}

// loadStorage creates a new Locale object at package level based on the Global variables settings.
// It's called automatically when trying to use Get or GetD methods.
func (c *config) loadDomain(domain string, force bool) *config {
	c.Lock()

	if c.storage == nil {
		c.storage = make(map[string]*Locale)
	}
	if c.storage[c.language] == nil || force {
		c.storage[c.language] = NewLocale(c.library, c.language)
	}
	if _, ok := c.storage[c.language].domains[c.domain]; !ok || force {
		c.storage[c.language].AddDomain(c.domain)
	}
	if _, ok := c.storage[c.language].domains[domain]; !ok || force {
		c.storage[c.language].AddDomain(domain)
	}

	c.Unlock()
	return c
}

// GetDomain is the domain getter for the package configuration
func (c *config) GetDomain() string {
	c.RLock()
	dom := c.domain
	c.RUnlock()

	return dom
}

// SetDomain sets the name for the domain to be used at package level.
// It reloads the corresponding translation file.
func (c *config) SetDomain(dom string) *config {
	c.Lock()
	c.domain = dom
	c.Unlock()

	c.loadStorage(true)
	return c
}

// GetLanguage is the language getter for the package configuration
func (c *config) GetLanguage() string {
	c.RLock()
	lang := c.language
	c.RUnlock()

	return lang
}

// SetLanguage sets the language code to be used at package level.
// It reloads the corresponding translation file.
func (c *config) SetLanguage(lang string) *config {
	c.Lock()
	c.language = lang
	c.Unlock()

	c.loadStorage(true)
	return c
}

// GetLibrary is the library getter for the package configuration
func (c *config) GetLibrary() string {
	c.RLock()
	lib := c.library
	c.RUnlock()

	return lib
}

// SetLibrary sets the root path for the loale directories and files to be used at package level.
// It reloads the corresponding translation file.
func (c *config) SetLibrary(lib string) *config {
	c.Lock()
	c.library = lib
	c.Unlock()

	c.loadStorage(true)
	return c
}

// Configure sets all configuration variables to be used at package level and reloads the corresponding translation file.
// It receives the library path, language code and domain name.
// This function is recommended to be used when changing more than one setting,
// as using each setter will introduce a I/O overhead because the translation file will be loaded after each set.
func (c *config) Configure(lib, lang, dom string) *config {
	c.Lock()

	c.library = lib
	c.language = lang
	c.domain = dom

	c.Unlock()

	c.loadStorage(true)
	return c
}

// Get uses the default domain globally set to return the corresponding translation of a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (c *config) Get(str string, vars ...interface{}) string {
	return c.GetD(c.GetDomain(), str, vars...)
}

// GetN retrieves the (N)th plural form of translation for the given string in the default domain.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (c *config) GetN(str, plural string, n int, vars ...interface{}) string {
	return c.GetND(c.GetDomain(), str, plural, n, vars...)
}

// GetD returns the corresponding translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (c *config) GetD(dom, str string, vars ...interface{}) string {
	return c.GetND(dom, str, str, 1, vars...)
}

// GetND retrieves the (N)th plural form of translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (c *config) GetND(dom, str, plural string, n int, vars ...interface{}) string {
	// Try to load default package Locale storage
	c.loadStorage(false)

	// Return translation
	c.RLock()
	tr := c.storage[c.language].GetND(dom, str, plural, n, vars...)
	c.RUnlock()

	return tr
}

// GetC uses the default domain globally set to return the corresponding translation of the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (c *config) GetC(str, ctx string, vars ...interface{}) string {
	return c.GetDC(c.GetDomain(), str, ctx, vars...)
}

// GetNC retrieves the (N)th plural form of translation for the given string in the given context in the default domain.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (c *config) GetNC(str, plural string, n int, ctx string, vars ...interface{}) string {
	return c.GetNDC(c.GetDomain(), str, plural, n, ctx, vars...)
}

// GetDC returns the corresponding translation in the given domain for the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (c *config) GetDC(dom, str, ctx string, vars ...interface{}) string {
	return c.GetNDC(dom, str, str, 1, ctx, vars...)
}

// GetNDC retrieves the (N)th plural form of translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (c *config) GetNDC(dom, str, plural string, n int, ctx string, vars ...interface{}) string {
	// Try to load default package Locale storage
	c.loadStorage(false)

	// Return translation
	c.RLock()
	tr := c.storage[c.language].GetNDC(dom, str, plural, n, ctx, vars...)
	c.RUnlock()

	return tr
}

// GetDomain is the domain getter for the package configuration
func GetDomain() string {
	globalConfig.RLock()
	dom := globalConfig.domain
	globalConfig.RUnlock()

	return dom
}

// SetDomain sets the name for the domain to be used at package level.
// It reloads the corresponding translation file.
func SetDomain(dom string) {
	globalConfig.Lock()
	globalConfig.domain = dom
	globalConfig.Unlock()

	globalConfig.loadStorage(true)
}

// GetLanguage is the language getter for the package configuration
func GetLanguage() string {
	globalConfig.RLock()
	lang := globalConfig.language
	globalConfig.RUnlock()

	return lang
}

// SetLanguage sets the language code to be used at package level.
// It reloads the corresponding translation file.
func SetLanguage(lang string) {
	globalConfig.Lock()
	globalConfig.language = lang
	globalConfig.Unlock()

	globalConfig.loadStorage(true)
}

// GetLibrary is the library getter for the package configuration
func GetLibrary() string {
	globalConfig.RLock()
	lib := globalConfig.library
	globalConfig.RUnlock()

	return lib
}

// SetLibrary sets the root path for the loale directories and files to be used at package level.
// It reloads the corresponding translation file.
func SetLibrary(lib string) {
	globalConfig.Lock()
	globalConfig.library = lib
	globalConfig.Unlock()

	globalConfig.loadStorage(true)
}

// Configure sets all configuration variables to be used at package level and reloads the corresponding translation file.
// It receives the library path, language code and domain name.
// This function is recommended to be used when changing more than one setting,
// as using each setter will introduce a I/O overhead because the translation file will be loaded after each set.
func Configure(lib, lang, dom string) {
	globalConfig.Lock()

	globalConfig.library = lib
	globalConfig.language = lang
	globalConfig.domain = dom

	globalConfig.Unlock()

	globalConfig.loadStorage(true)
}

// Get uses the default domain globally set to return the corresponding translation of a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func Get(str string, vars ...interface{}) string {
	return GetD(GetDomain(), str, vars...)
}

// GetN retrieves the (N)th plural form of translation for the given string in the default domain.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetN(str, plural string, n int, vars ...interface{}) string {
	return GetND(GetDomain(), str, plural, n, vars...)
}

// GetD returns the corresponding translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetD(dom, str string, vars ...interface{}) string {
	return GetND(dom, str, str, 1, vars...)
}

// GetND retrieves the (N)th plural form of translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetND(dom, str, plural string, n int, vars ...interface{}) string {
	return globalConfig.GetND(dom, str, plural, n, vars...)
}

// GetC uses the default domain globally set to return the corresponding translation of the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetC(str, ctx string, vars ...interface{}) string {
	return GetDC(GetDomain(), str, ctx, vars...)
}

// GetNC retrieves the (N)th plural form of translation for the given string in the given context in the default domain.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetNC(str, plural string, n int, ctx string, vars ...interface{}) string {
	return GetNDC(GetDomain(), str, plural, n, ctx, vars...)
}

// GetDC returns the corresponding translation in the given domain for the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetDC(dom, str, ctx string, vars ...interface{}) string {
	return GetNDC(dom, str, str, 1, ctx, vars...)
}

// GetNDC retrieves the (N)th plural form of translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetNDC(dom, str, plural string, n int, ctx string, vars ...interface{}) string {
	return globalConfig.GetNDC(dom, str, plural, n, ctx, vars...)
}

// printf applies text formatting only when needed to parse variables.
func printf(str string, vars ...interface{}) string {
	if len(vars) > 0 {
		return fmt.Sprintf(str, vars...)
	}

	return str
}
