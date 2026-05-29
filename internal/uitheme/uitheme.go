// Package uitheme provides a Fyne theme wrapper that applies user-selected
// font size, layout density and font family on top of any base theme.
package uitheme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme wraps a base Fyne theme and overrides Size and Font.
type CustomTheme struct {
	base     fyne.Theme
	fontSize float32
	density  string
	fontName string
}

// New creates a CustomTheme.
//   - base     : Light/Dark/Default theme from fyne/theme
//   - fontSize : user-selected text size in logical pixels (10-24)
//   - density  : "Compact" | "Normal" | "Comfortable"
//   - fontName : "System" | "Monospace"
func New(base fyne.Theme, fontSize float32, density, fontName string) *CustomTheme {
	if base == nil {
		base = theme.DefaultTheme()
	}
	return &CustomTheme{
		base:     base,
		fontSize: fontSize,
		density:  density,
		fontName: fontName,
	}
}

// Color delegates to the base theme.
func (t *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}

// Icon delegates to the base theme.
func (t *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

// Font returns the monospace font when the user chose Monospace or when the
// text style itself requests it; otherwise delegates to the base theme.
func (t *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	if t.fontName == "Monospace" || style.Monospace {
		return theme.DefaultTextMonospaceFont()
	}
	return t.base.Font(style)
}

// Size overrides text sizes with the user-selected value and scales padding
// according to the chosen density.
func (t *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	df := t.densityFactor()
	switch name {
	case theme.SizeNameText:
		return t.fontSize
	case theme.SizeNameCaptionText:
		return max32(t.fontSize*0.8, 9)
	case theme.SizeNameHeadingText:
		return t.fontSize * 1.5
	case theme.SizeNameSubHeadingText:
		return t.fontSize * 1.2
	case theme.SizeNamePadding:
		return t.base.Size(name) * df
	case theme.SizeNameInnerPadding:
		return t.base.Size(name) * df
	case theme.SizeNameLineSpacing:
		return t.base.Size(name) * df
	default:
		return t.base.Size(name)
	}
}

func (t *CustomTheme) densityFactor() float32 {
	switch t.density {
	case "Compact":
		return 0.65
	case "Comfortable":
		return 1.35
	default: // Normal
		return 1.0
	}
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
