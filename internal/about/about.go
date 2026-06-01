package about

import (
	"fmt"
	"image/color"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"msxdbdown/internal/appicon"
)

// Show displays the About dialog with version and copyright information.
//
//   - parent: main application window
//   - version: e.g. "0.1.7"
//   - buildDate: military date, e.g. "29052026"
//   - buildTime: military time, e.g. "1433"
//   - buildNumber: hex timestamp, e.g. "67a4f8c2"
//   - tr: i18n translation function (key -> localized string)
func Show(parent fyne.Window, version, buildDate, buildTime, buildNumber string, tr func(string) string) {
	// Title: "MSX DB Down" in larger, bold, blue text
	icon := canvas.NewImageFromResource(appicon.Resource())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(84, 84))

	titleText := canvas.NewText("MSX DB Down", color.NRGBA{0x00, 0x66, 0xCC, 0xFF})
	titleText.TextSize = 24
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	// Helper function to create centered bold labels.
	centerBoldLabel := func(text string) *widget.Label {
		return widget.NewLabelWithStyle(text, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	}

	versionBuildLine := centerBoldLabel(fmt.Sprintf("%s: %s  •  %s: %s", tr("aboutVersion"), version, tr("aboutBuild"), buildNumber))
	dateTimeLine := centerBoldLabel(fmt.Sprintf("%s: %s  •  %s: %s", tr("aboutDate"), buildDate, tr("aboutTime"), buildTime))
	copyrightLine := widget.NewLabelWithStyle(
		fmt.Sprintf("%s - %s", tr("aboutCopyright1"), tr("aboutCopyright2")),
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)

	// Website as clickable hyperlink
	website := "www.cybernostra.com"
	websiteURL, _ := url.Parse("https://www.cybernostra.com")
	websiteLink := widget.NewHyperlink(website, websiteURL)
	websiteLink.Alignment = fyne.TextAlignCenter

	// Years
	yearsLabel := widget.NewLabel(tr("aboutYears"))
	yearsLabel.Alignment = fyne.TextAlignCenter

	// Separator
	sep1 := widget.NewSeparator()
	sep2 := widget.NewSeparator()

	// Assemble content
	content := container.NewVBox(
		container.NewCenter(icon),
		titleText,
		sep1,
		versionBuildLine,
		dateTimeLine,
		sep2,
		copyrightLine,
		websiteLink,
		yearsLabel,
	)

	// Wrap in a padded container
	paddedContent := container.NewPadded(content)

	// Show dialog with larger size
	dlg := dialog.NewCustom(tr("aboutTitle"), tr("aboutClose"), paddedContent, parent)
	dlg.Resize(fyne.NewSize(500, 300))
	dlg.Show()
}
