// Package configui provides the "Config UI" settings dialog.
package configui

import (
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"msxdbdown/internal/uiprefs"
)

const (
	catalogDBLocal   = "local"
	catalogDBAppData = "appdata"
)

// Show opens the Config UI modal dialog.
//
//   - parent  : main window (needed by Fyne dialogs)
//   - getValue: read a persisted value by key
//   - setValue: write a persisted value by key
//   - tr      : i18n translation function (key -> localized string)
//   - onApply : callback invoked after the user confirms; caller should
//     re-apply the combined theme so changes take effect immediately
func Show(
	parent fyne.Window,
	getValue func(string) string,
	setValue func(string, string),
	currentDBLocation func() string,
	currentDBPath func() string,
	resolveDBPath func(string) (string, error),
	applyDBLocation func(targetLocation string, moveCurrent bool) error,
	tr func(string) string,
	onApply func(),
) {
	// ---- read current values ------------------------------------------------
	currentSize := uiprefs.ReadFontSize(getValue(uiprefs.PrefFontSize))
	currentDensity := uiprefs.ReadDensity(getValue(uiprefs.PrefDensity))
	currentFont := uiprefs.ReadFontName(getValue(uiprefs.PrefFontName))
	currentMSXRomDBURL := uiprefs.ReadURL(getValue(uiprefs.PrefMSXRomDBURL), uiprefs.DefaultMSXRomDBURL)
	currentFileHunterURL := uiprefs.ReadURL(getValue(uiprefs.PrefFileHunterURL), uiprefs.DefaultFileHunterURL)
	currentFileHunterSHAURL := uiprefs.ReadURL(getValue(uiprefs.PrefFileHunterSHAURL), uiprefs.DefaultFileHunterSHAURL)
	selectedCatalogDBLocation := uiprefs.ReadCatalogDBLocation(currentDBLocation())

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

	// ---- database source URLs -----------------------------------------------
	msxRomDBEntry := widget.NewEntry()
	msxRomDBEntry.SetText(currentMSXRomDBURL)

	fileHunterEntry := widget.NewEntry()
	fileHunterEntry.SetText(currentFileHunterURL)

	fileHunterSHAEntry := widget.NewEntry()
	fileHunterSHAEntry.SetText(currentFileHunterSHAURL)

	// ---- UI tab -------------------------------------------------------------
	uiTab := container.NewVBox(
		widget.NewLabelWithStyle(tr("configFontFamily"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		fontSelect,
		widget.NewSeparator(),
		widget.NewLabelWithStyle(tr("configFontSize"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		sizeRow,
		widget.NewSeparator(),
		widget.NewLabelWithStyle(tr("configDensity"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		densitySelect,
	)

	// ---- URLs tab -----------------------------------------------------------
	urlTab := container.NewVBox(
		widget.NewLabelWithStyle(tr("configMSXRomDBURL"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		msxRomDBEntry,
		widget.NewSeparator(),
		widget.NewLabelWithStyle(tr("configFileHunterURL"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		fileHunterEntry,
		widget.NewSeparator(),
		widget.NewLabelWithStyle(tr("configFileHunterSHAURL"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		fileHunterSHAEntry,
	)

	// ---- SQLite tab ---------------------------------------------------------
	locationSelect := widget.NewRadioGroup(
		[]string{tr("configDBLocationLocal"), tr("configDBLocationAppData")},
		nil,
	)
	locationFromSelection := func(selected string) string {
		if selected == tr("configDBLocationAppData") {
			return catalogDBAppData
		}
		return catalogDBLocal
	}
	selectionFromLocation := func(location string) string {
		if uiprefs.ReadCatalogDBLocation(location) == catalogDBAppData {
			return tr("configDBLocationAppData")
		}
		return tr("configDBLocationLocal")
	}
	locationSelect.SetSelected(selectionFromLocation(selectedCatalogDBLocation))

	currentDBPathLabel := widget.NewLabel(currentDBPath())
	currentDBPathLabel.Wrapping = fyne.TextWrapWord

	dbPathLabel := widget.NewLabel("")
	dbPathLabel.Wrapping = fyne.TextWrapWord
	updateDBPathLabel := func() {
		dbPath, err := resolveDBPath(locationFromSelection(locationSelect.Selected))
		if err != nil {
			dbPathLabel.SetText(fmt.Sprintf("%s: %v", tr("configDBPathError"), err))
			return
		}
		dbPathLabel.SetText(dbPath)
	}
	locationSelect.OnChanged = func(string) {
		updateDBPathLabel()
	}
	updateDBPathLabel()

	createDBBtn := widget.NewButton(tr("configCreateCatalogDB"), func() {
		selectedLocation := locationFromSelection(locationSelect.Selected)
		dbPath, err := resolveDBPath(selectedLocation)
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}
		currentLocation := uiprefs.ReadCatalogDBLocation(currentDBLocation())

		createFresh := func(moveCurrent bool) {
			if err := applyDBLocation(selectedLocation, moveCurrent); err != nil {
				dialog.ShowError(err, parent)
				return
			}
			selectedCatalogDBLocation = uiprefs.ReadCatalogDBLocation(selectedLocation)
			locationSelect.SetSelected(selectionFromLocation(selectedCatalogDBLocation))
			currentDBPathLabel.SetText(currentDBPath())
			updateDBPathLabel()
			onApply()
			dialog.ShowInformation(tr("configDBCreatedTitle"), fmt.Sprintf(tr("configDBCreatedMessage"), currentDBPath()), parent)
		}

		if selectedLocation != currentLocation {
			dialog.ShowConfirm(
				tr("configDBSwitchTitle"),
				fmt.Sprintf(tr("configDBSwitchAskMove"), currentDBPath(), dbPath),
				func(moveCurrent bool) {
					if moveCurrent {
						createFresh(true)
						return
					}
					dialog.ShowConfirm(
						tr("configDBSwitchTitle"),
						fmt.Sprintf(tr("configDBSwitchAskNew"), dbPath),
						func(createNew bool) {
							if createNew {
								createFresh(false)
							}
						},
						parent,
					)
				},
				parent,
			)
			return
		}

		if _, statErr := os.Stat(dbPath); statErr == nil {
			dialog.ShowConfirm(
				tr("configDBExistsTitle"),
				fmt.Sprintf(tr("configDBExistsConfirm"), dbPath),
				func(confirm bool) {
					if confirm {
						createFresh(false)
					}
				},
				parent,
			)
			return
		}

		createFresh(false)
	})

	sqliteTab := container.NewVBox(
		widget.NewLabelWithStyle(tr("configCurrentDBPath"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		currentDBPathLabel,
		widget.NewSeparator(),
		widget.NewLabelWithStyle(tr("configCatalogDBLocation"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		locationSelect,
		widget.NewSeparator(),
		widget.NewLabelWithStyle(tr("configCatalogDBPath"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		dbPathLabel,
		widget.NewSeparator(),
		createDBBtn,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem(tr("configTabUI"), container.NewPadded(uiTab)),
		container.NewTabItem(tr("configTabURLs"), container.NewPadded(urlTab)),
		container.NewTabItem(tr("configTabSQLite"), container.NewPadded(sqliteTab)),
	)

	// ---- dialog -------------------------------------------------------------
	dlg := dialog.NewCustomConfirm(
		tr("configTitle"),
		tr("configOK"),
		tr("configCancel"),
		container.NewPadded(tabs),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			setValue(uiprefs.PrefFontSize, strconv.Itoa(int(sizeSlider.Value)))
			setValue(uiprefs.PrefDensity, densitySelect.Selected)
			setValue(uiprefs.PrefFontName, fontSelect.Selected)
			setValue(uiprefs.PrefMSXRomDBURL, msxRomDBEntry.Text)
			setValue(uiprefs.PrefFileHunterURL, fileHunterEntry.Text)
			setValue(uiprefs.PrefFileHunterSHAURL, fileHunterSHAEntry.Text)
			onApply()
		},
		parent,
	)
	dlg.Resize(fyne.NewSize(700, 620))
	dlg.Show()
}
