// Package configui provides the "Config UI" settings dialog.
package configui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"msxdbdown/internal/uiprefs"
)

// Show opens the Config UI modal dialog.
//
//   - parent  : main window (needed by Fyne dialogs)
//   - getValue: read a persisted value by key
//   - setValue: write a persisted value by key
//   - tr      : i18n translation function (key → localized string)
//   - onApply : callback invoked after the user confirms; caller should
//     re-apply the combined theme so changes take effect immediately
func Show(
	parent fyne.Window,
	getValue func(string) string,
	setValue func(string, string),
	tr func(string) string,
	onApply func(),
) {
	// ---- read current values ------------------------------------------------
	currentSize := uiprefs.ReadFontSize(getValue(uiprefs.PrefFontSize))
	currentDensity := uiprefs.ReadDensity(getValue(uiprefs.PrefDensity))
	currentFont := uiprefs.ReadFontName(getValue(uiprefs.PrefFontName))

	// ---- font size ----------------------------------------------------------
	sizeValueLabel := widget.NewLabel(fmt.Sprintf(tr("configFontSizeValue"), int(currentSize)))

	sizeSlider := widget.NewSlider(10, 24)
	sizeSlider.Step = 1
	sizeSlider.Value = float64(currentSize)
	sizeSlider.OnChanged = func(v float64) {
		sizeValueLabel.SetText(fmt.Sprintf(tr("configFontSizeValue"), int(v)))
	}

	sizeRow := container.NewBorder(nil, nil, nil, sizeValueLabel, sizeSlider)

	// ---- font family --------------------------------------------------------
	fontSelect := widget.NewSelect(uiprefs.FontNameOptions(), nil)
	fontSelect.SetSelected(currentFont)

	// ---- layout density -----------------------------------------------------
	densitySelect := widget.NewSelect(uiprefs.DensityOptions(), nil)
	densitySelect.SetSelected(currentDensity)

	// ---- assemble form ------------------------------------------------------
	form := container.NewVBox(
		widget.NewLabelWithStyle(tr("configFontFamily"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		fontSelect,

		widget.NewSeparator(),

		widget.NewLabelWithStyle(tr("configFontSize"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		sizeRow,

		widget.NewSeparator(),

		widget.NewLabelWithStyle(tr("configDensity"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		densitySelect,
	)

	// ---- dialog -------------------------------------------------------------
	dlg := dialog.NewCustomConfirm(
		tr("configTitle"),
		tr("configOK"),
		tr("configCancel"),
		container.NewPadded(form),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			setValue(uiprefs.PrefFontSize, strconv.Itoa(int(sizeSlider.Value)))
			setValue(uiprefs.PrefDensity, densitySelect.Selected)
			setValue(uiprefs.PrefFontName, fontSelect.Selected)
			onApply()
		},
		parent,
	)
	dlg.Resize(fyne.NewSize(460, 380))
	dlg.Show()
}
