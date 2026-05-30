package uiprefs

import (
	"strconv"
	"strings"
)

const (
	PrefMSXRomDBURL       = "db.msxromdb.url"
	PrefFileHunterURL     = "db.filehunter.url"
	PrefFileHunterSHAURL  = "db.filehunter.sha.url"
	PrefCatalogDBLocation = "db.catalog.location"

	DefaultMSXRomDBURL       = "https://romdb.vampier.net/Archive/sql-msxromdb.zip"
	DefaultFileHunterURL     = "https://download.file-hunter.com/allfiles.txt"
	DefaultFileHunterSHAURL  = "https://download.file-hunter.com/sha1sums.txt"
	DefaultCatalogDBLocation = "local"
)

func ReadCatalogDBLocation(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "appdata":
		return "appdata"
	default:
		return DefaultCatalogDBLocation
	}
}

// ---------- language ----------------------------------------------------------

type LanguageCode string

const (
	LangPT LanguageCode = "pt"
	LangEN LanguageCode = "en"
	LangES LanguageCode = "es"
	LangNL LanguageCode = "nl"
	LangIT LanguageCode = "it"
)

var LanguageDisplay = map[LanguageCode]string{
	LangPT: "Português",
	LangEN: "English",
	LangES: "Español",
	LangNL: "Nederlands",
	LangIT: "Italiano",
}

var DisplayToLanguage = map[string]LanguageCode{
	"Português":  LangPT,
	"English":    LangEN,
	"Español":    LangES,
	"Nederlands": LangNL,
	"Italiano":   LangIT,
}

func ReadLanguage(value string) LanguageCode {
	lang := LanguageCode(value)
	if _, ok := LanguageDisplay[lang]; ok {
		return lang
	}
	return LangEN
}

func ReadTheme(value string) string {
	switch strings.ToLower(value) {
	case "light":
		return "Light"
	case "dark":
		return "Dark"
	default:
		return "System"
	}
}

func LanguageOptions() []string {
	return []string{
		LanguageDisplay[LangPT],
		LanguageDisplay[LangEN],
		LanguageDisplay[LangES],
		LanguageDisplay[LangNL],
		LanguageDisplay[LangIT],
	}
}

// ---------- font & density ----------------------------------------------------

const (
	PrefFontName    = "ui.fontName"
	PrefFontSize    = "ui.fontSize"
	PrefDensity     = "ui.density"
	DefaultFontSize = float32(14)
	DefaultDensity  = "Normal"
	DefaultFontName = "System"
)

func FontSizeOptions() []string {
	return []string{"10", "11", "12", "13", "14", "15", "16", "18", "20", "22", "24"}
}

func DensityOptions() []string {
	return []string{"Compact", "Normal", "Comfortable"}
}

func FontNameOptions() []string {
	return []string{"System", "Monospace"}
}

// ReadFontSize parses the stored string; returns DefaultFontSize on any error.
func ReadFontSize(value string) float32 {
	if value == "" {
		return DefaultFontSize
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(value), 32)
	if err != nil || f < 8 || f > 40 {
		return DefaultFontSize
	}
	return float32(f)
}

// ReadDensity validates the stored density; returns DefaultDensity if unknown.
func ReadDensity(value string) string {
	for _, opt := range DensityOptions() {
		if strings.EqualFold(value, opt) {
			return opt
		}
	}
	return DefaultDensity
}

// ReadFontName validates the stored font name; returns DefaultFontName if unknown.
func ReadFontName(value string) string {
	for _, opt := range FontNameOptions() {
		if strings.EqualFold(value, opt) {
			return opt
		}
	}
	return DefaultFontName
}

func ReadURL(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
