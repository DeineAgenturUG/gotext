/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */
/*
Package gnuGettext implements GNU gettext like utilities.

For quick/simple translations you can use the package level functions directly.

    import (
	    "fmt"
	    "github.com/leonelquinteros/gotext/gnuGettext"
    )

    func main() {
        // Configure package
        gnuGettext.Configure("/path/to/locales/root/dir", "en_UK", "domain-name")

        // Translate text from default domain
        fmt.Println(gnuGettext.Gettext("My text on 'domain-name' domain"))

        // Translate text from a different domain without reconfigure
        fmt.Println(gnuGettext.Dgettext("domain2", "Another text on a different domain"))
    }

*/
package gnuGettext

import (
	"github.com/leonelquinteros/gotext"
	"github.com/leonelquinteros/gotext/format"
)

// Sprinf alias from format submodule
var Sprint = format.Sprintf

func Gettext(msgid string) string {
	return gotext.Get(msgid)
}
func Dgettext(domain, msgid string) string {
	return gotext.GetD(domain, msgid)
}
func Ngettext(msgid, msgidPlural string, count int) string {
	return gotext.GetN(msgid, msgidPlural, count)
}
func Dngettext(domain, msgid, msgidPlural string, count int) string {
	return gotext.GetND(domain, msgid, msgidPlural, count)
}
func Pgettext(msgctxt, msgid string) string {
	return gotext.GetC(msgid, msgctxt)
}
func Dpgettext(domain, msgctxt, msgid string) string {
	return gotext.GetDC(domain, msgid, msgctxt)
}
func Npgettext(msgctxt, msgid, msgidPlural string, count int) string {
	return gotext.GetNC(msgid, msgidPlural, count, msgctxt)
}
func Dnpgettext(domain, msgctxt, msgid, msgidPlural string, count int) string {
	return gotext.GetNDC(domain, msgid, msgidPlural, count, msgctxt)
}
func SetLocale(locale string) {
	gotext.SetLanguage(locale)
}
func GetLocale() string {
	return gotext.GetLanguage()
}
func SetTextDomain(domain string) {
	gotext.SetDomain(domain)
}
func GetTextDomain() string {
	return gotext.GetDomain()
}
func SetLibary(libary string) {
	gotext.SetLibrary(libary)
}
func GetLibary() string {
	return gotext.GetLibrary()
}
