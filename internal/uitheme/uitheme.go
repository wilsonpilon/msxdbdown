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
	bg := toNRGBA(t.base.Color(theme.ColorNameBackground, variant))
	isDarkBG := relativeLuma(bg) < 0.5

	switch name {
	case theme.ColorNameForeground, theme.ColorNamePlaceHolder, theme.ColorNameDisabled:
		if isDarkBG {
			return color.NRGBA{R: 0xF2, G: 0xF4, B: 0xF8, A: 0xFF}
		}
		return color.NRGBA{R: 0x12, G: 0x16, B: 0x1D, A: 0xFF}
	case theme.ColorNameInputBackground:
		if isDarkBG {
			return color.NRGBA{R: 0x2A, G: 0x30, B: 0x39, A: 0xFF}
		}
		return color.NRGBA{R: 0xEE, G: 0xF1, B: 0xF6, A: 0xFF}
	}

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

func toNRGBA(c color.Color) color.NRGBA {
	if v, ok := c.(color.NRGBA); ok {
		return v
	}
	r, g, b, a := c.RGBA()
	return color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

func relativeLuma(c color.NRGBA) float64 {
	return (0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)) / 255.0
}
