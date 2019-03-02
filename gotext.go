/*
Package gotext implements GNU gettext utilities.

For quick/simple translations you can use the package level functions directly.

    import (
	    "fmt"
	    "github.com/DeineAgenturUG/gotext"
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
	"encoding/gob"
	"sync"
)

// Global environment variables
type config struct {
	sync.RWMutex

	// Default domain to look at when no domain is specified. Used by package level functions.
	loadDomains []string
	domain      string

	// Language set.
	loadLanguages []string
	language      string

	// Path to library directory where all locale directories and Translation files are.
	library string

	// Storage for package level methods
	storage sync.Map
}

var (
	globalConfig *config
	once         sync.Once

	// DefaultDomain as mostly used
	DefaultDomain = "default"
)

func init() {
	// Init default configuration
	globalConfig = &config{
		loadDomains:   []string{"default"},
		domain:        "default",
		loadLanguages: []string{"en_US"},
		language:      "en_US",
		library:       "/usr/local/share/locale",
		storage:       sync.Map{},
	}

	// Register Translator types for gob encoding
	gob.Register(TranslatorEncoding{})
}

// GetInstance Create Instance default configuration
func GetInstance(loadDomains, loadLanguages []string, defaultDomain, defaultLanguage, library string) {
	once.Do(func() {
		globalConfig = &config{
			loadDomains:   loadDomains,
			domain:        defaultDomain,
			loadLanguages: loadLanguages,
			language:      defaultLanguage,
			library:       library,
			storage:       sync.Map{},
		}
		globalConfig.loadStorage(true)
	})
}

// loadStorage creates a new Locale object at package level based on the Global variables settings.
// It's called automatically when trying to use Get or GetD methods.
func (c *config) loadStorage(force bool) *config {
	c.RLock()

	if v, _ := c.storage.LoadOrStore(c.language, NewLocale(c.library, c.language)); v != nil {
		v2 := v.(*Locale)
		v2.AddDomain(c.domain)
		for _, domain := range c.loadDomains {
			if _, ok := v2.Domains.Load(domain); !ok || force {
				v2.AddDomain(domain)
			}
		}
	}

	for _, language := range c.loadLanguages {
		if v, _ := c.storage.LoadOrStore(language, NewLocale(c.library, language)); v != nil {
			v2 := v.(*Locale)
			v2.AddDomain(c.domain)
			for _, domain := range c.loadDomains {
				if _, ok := v2.Domains.Load(domain); !ok || force {
					v2.AddDomain(domain)
				}
			}
		}
	}
	c.RUnlock()
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
	c.loadDomains = append(c.loadDomains, dom)
	c.loadDomains = UniqStrings(c.loadDomains)
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

	var tr string

	globalConfig.RLock()
	var local = c.language
	globalConfig.RUnlock()

	if v, _ := c.storage.Load(local); v != nil {
		v2 := v.(*Locale)
		var notIncluded = true
		v2.Domains.Range(func(k, v interface{}) bool {
			if k == dom {
				notIncluded = false
			}
			return true
		})
		if notIncluded {
			v2.AddDomain(dom)
		}
		c.loadStorage(true)
		tr = v2.GetND(dom, str, plural, n, vars...)
	}

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
	var tr string
	globalConfig.RLock()
	var local = c.language
	globalConfig.RUnlock()
	if v, _ := c.storage.Load(local); v != nil {
		v2 := v.(*Locale)
		var notIncluded = true
		v2.Domains.Range(func(key, value interface{}) bool {
			if key == dom {
				notIncluded = false
			}
			return true
		})
		if notIncluded {
			v2.AddDomain(dom)
		}
		c.loadStorage(true)
		tr = v2.GetNDC(dom, str, plural, n, ctx, vars...)
	}

	return tr
}

// GetDomain is the domain getter for the package configuration
func GetDomain() string {
	var dom string

	if v, _ := globalConfig.storage.Load(globalConfig.GetLanguage()); v != nil {
		v2 := v.(*Locale)
		dom = v2.GetDomain()
	}

	globalConfig.RLock()
	if dom == "" {
		dom = globalConfig.domain
	}
	globalConfig.RUnlock()

	return dom
}

// SetDomain sets the name for the domain to be used at package level.
// It reloads the corresponding Translation file.
func SetDomain(dom string) {
	globalConfig.Lock()
	globalConfig.domain = dom
	globalConfig.Unlock()

	globalConfig.storage.Range(func(key interface{}, value interface{}) bool {
		storage := value.(*Locale)
		storage.SetDomain(dom)
		return true
	})

	globalConfig.loadStorage(true)
}

// GetLanguage is the language getter for the package configuration
func GetLanguage() string {
	globalConfig.RLock()
	defer globalConfig.RUnlock()

	return globalConfig.language
}

// SetLanguage sets the language code to be used at package level.
// It reloads the corresponding Translation file.
func SetLanguage(lang string) {
	globalConfig.Lock()
	globalConfig.language = SimplifiedLocale(lang)
	globalConfig.Unlock()

	globalConfig.loadStorage(true)
}

// GetLibrary is the library getter for the package configuration
func GetLibrary() string {
	globalConfig.RLock()
	defer globalConfig.RUnlock()

	return globalConfig.library
}

// SetLibrary sets the root path for the loale directories and files to be used at package level.
// It reloads the corresponding Translation file.
func SetLibrary(lib string) {
	globalConfig.Lock()
	globalConfig.library = lib
	globalConfig.Unlock()

	globalConfig.loadStorage(true)
}

// Configure sets all configuration variables to be used at package level and reloads the corresponding Translation file.
// It receives the library path, language code and domain name.
// This function is recommended to be used when changing more than one setting,
// as using each setter will introduce a I/O overhead because the Translation file will be loaded after each set.
func Configure(lib, lang, dom string) {
	globalConfig.Lock()
	globalConfig.library = lib
	globalConfig.language = SimplifiedLocale(lang)
	globalConfig.domain = dom
	globalConfig.loadDomains = append(globalConfig.loadDomains, dom)
	globalConfig.loadDomains = UniqStrings(globalConfig.loadDomains)
	globalConfig.Unlock()

	globalConfig.loadStorage(true)
}

// Get uses the default domain globally set to return the corresponding Translation of a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func Get(str string, vars ...interface{}) string {
	return GetD(GetDomain(), str, vars...)
}

// GetN retrieves the (N)th plural form of Translation for the given string in the default domain.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetN(str, plural string, n int, vars ...interface{}) string {
	return GetND(GetDomain(), str, plural, n, vars...)
}

// GetD returns the corresponding Translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetD(dom, str string, vars ...interface{}) string {
	return GetND(dom, str, str, 1, vars...)
}

// GetND retrieves the (N)th plural form of Translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetND(dom, str, plural string, n int, vars ...interface{}) string {
	return globalConfig.GetND(dom, str, plural, n, vars...)
}

// GetC uses the default domain globally set to return the corresponding Translation of the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetC(str, ctx string, vars ...interface{}) string {
	return GetDC(GetDomain(), str, ctx, vars...)
}

// GetNC retrieves the (N)th plural form of Translation for the given string in the given context in the default domain.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetNC(str, plural string, n int, ctx string, vars ...interface{}) string {
	return GetNDC(GetDomain(), str, plural, n, ctx, vars...)
}

// GetDC returns the corresponding Translation in the given domain for the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetDC(dom, str, ctx string, vars ...interface{}) string {
	return GetNDC(dom, str, str, 1, ctx, vars...)
}

// GetNDC retrieves the (N)th plural form of Translation in the given domain for a given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func GetNDC(dom, str, plural string, n int, ctx string, vars ...interface{}) string {
	return globalConfig.GetNDC(dom, str, plural, n, ctx, vars...)
}
