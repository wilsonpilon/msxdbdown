package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/spf13/cobra"
	"msxdbdown/internal/about"
	"msxdbdown/internal/configui"
	"msxdbdown/internal/settingsdb"
	"msxdbdown/internal/uiprefs"
	"msxdbdown/internal/uitheme"
)

const (
	prefLanguage = "ui.language"
	prefTheme    = "ui.theme"
)

var i18n = map[uiprefs.LanguageCode]map[string]string{
	uiprefs.LangPT: {
		"appTitle":            "MSX DB Down - Catálogo",
		"menuFile":            "Arquivo",
		"menuExit":            "Sair",
		"menuSetup":           "Configuração",
		"menuConfigUI":        "Configurar UI",
		"menuHelp":            "Ajuda",
		"menuAbout":           "Sobre",
		"language":            "Idioma",
		"theme":               "Tema",
		"panelControls":       "Preferências",
		"panelPreview":        "Área principal",
		"previewText":         "Catálogo visual (em breve)\n\nAqui serão exibidos jogos, músicas, imagens e metadados.",
		"statusReady":         "Pronto.",
		"statusLogTitle":      "Status e depuração",
		"configTitle":         "Configurar Interface",
		"configFontFamily":    "Família de Fonte",
		"configFontSize":      "Tamanho de Fonte",
		"configFontSizeValue": "%d px",
		"configDensity":       "Densidade de Layout",
		"configOK":            "Aplicar",
		"configCancel":        "Cancelar",
		"aboutTitle":          "Sobre",
		"aboutVersion":        "Versão",
		"aboutBuild":          "Build",
		"aboutDate":           "Data",
		"aboutTime":           "Hora",
		"aboutCopyright1":     "(C) WIB Projetos Ltda.",
		"aboutCopyright2":     "(C) Cybernostra, Inc.",
		"aboutWebsite":        "www.cybernostra.com",
		"aboutYears":          "1972 - %d",
		"aboutClose":          "Fechar",
		"cliShort":            "Baixador de banco de dados MSX e catalogo visual",
		"cliLong":             "MSX DB Down baixa bancos de software MSX, enriquece itens com metadados (imagens, musica, video, informacoes de lancamento) e fornece um catalogo visual como frontend para emuladores MSX.\n\nExecutar sem subcomando abre a interface grafica.",
		"cliFlagLang":         "Define idioma da UI: pt | en | es | nl | it",
		"cliFlagTheme":        "Define tema da UI: system | light | dark",
		"cliFlagDebug":        "Mostra mensagens extras de depuracao no painel de log",
		"cliVersionShort":     "Mostra informacoes de versao",
	},
	uiprefs.LangEN: {
		"appTitle":            "MSX DB Down - Catalog",
		"menuFile":            "File",
		"menuExit":            "Exit",
		"menuSetup":           "Setup",
		"menuConfigUI":        "Config UI",
		"menuHelp":            "Help",
		"menuAbout":           "About",
		"language":            "Language",
		"theme":               "Theme",
		"panelControls":       "Preferences",
		"panelPreview":        "Main area",
		"previewText":         "Visual catalog (coming soon)\n\nGames, music, images and metadata will be shown here.",
		"statusReady":         "Ready.",
		"statusLogTitle":      "Status and debug",
		"configTitle":         "Config UI",
		"configFontFamily":    "Font Family",
		"configFontSize":      "Font Size",
		"configFontSizeValue": "%d px",
		"configDensity":       "Layout Density",
		"configOK":            "Apply",
		"configCancel":        "Cancel",
		"aboutTitle":          "About",
		"aboutVersion":        "Version",
		"aboutBuild":          "Build",
		"aboutDate":           "Date",
		"aboutTime":           "Time",
		"aboutCopyright1":     "(C) WIB Projetos Ltda.",
		"aboutCopyright2":     "(C) Cybernostra, Inc.",
		"aboutWebsite":        "www.cybernostra.com",
		"aboutYears":          "1972 - %d",
		"aboutClose":          "Close",
		"cliShort":            "MSX Database Downloader and Visual Catalog",
		"cliLong":             "MSX DB Down downloads MSX software databases, enriches entries with metadata (images, music, video, release info) and provides a visual catalog that acts as a frontend for MSX emulators.\n\nRunning without a sub-command opens the graphical interface.",
		"cliFlagLang":         "Override UI language: pt | en | es | nl | it",
		"cliFlagTheme":        "Override UI theme: system | light | dark",
		"cliFlagDebug":        "Print extra debug messages in the log panel",
		"cliVersionShort":     "Print version information",
	},
	uiprefs.LangES: {
		"appTitle":            "MSX DB Down - Catálogo",
		"menuFile":            "Archivo",
		"menuExit":            "Salir",
		"menuSetup":           "Configuración",
		"menuConfigUI":        "Configurar UI",
		"menuHelp":            "Ayuda",
		"menuAbout":           "Acerca de",
		"language":            "Idioma",
		"theme":               "Tema",
		"panelControls":       "Preferencias",
		"panelPreview":        "Área principal",
		"previewText":         "Catálogo visual (próximamente)\n\nAquí se mostrarán juegos, música, imágenes y metadatos.",
		"statusReady":         "Listo.",
		"statusLogTitle":      "Estado y depuración",
		"configTitle":         "Configurar interfaz",
		"configFontFamily":    "Familia de fuente",
		"configFontSize":      "Tamaño de fuente",
		"configFontSizeValue": "%d px",
		"configDensity":       "Densidad de diseño",
		"configOK":            "Aplicar",
		"configCancel":        "Cancelar",
		"aboutTitle":          "Acerca de",
		"aboutVersion":        "Versión",
		"aboutBuild":          "Build",
		"aboutDate":           "Fecha",
		"aboutTime":           "Hora",
		"aboutCopyright1":     "(C) WIB Projetos Ltda.",
		"aboutCopyright2":     "(C) Cybernostra, Inc.",
		"aboutWebsite":        "www.cybernostra.com",
		"aboutYears":          "1972 - %d",
		"aboutClose":          "Cerrar",
		"cliShort":            "Descargador de base de datos MSX y catalogo visual",
		"cliLong":             "MSX DB Down descarga bases de software MSX, enriquece elementos con metadatos (imagenes, musica, video, informacion de lanzamiento) y ofrece un catalogo visual como frontend para emuladores MSX.\n\nEjecutar sin subcomando abre la interfaz grafica.",
		"cliFlagLang":         "Define idioma de la UI: pt | en | es | nl | it",
		"cliFlagTheme":        "Define tema de la UI: system | light | dark",
		"cliFlagDebug":        "Muestra mensajes extra de depuracion en el panel de log",
		"cliVersionShort":     "Muestra informacion de version",
	},
	uiprefs.LangNL: {
		"appTitle":            "MSX DB Down - Catalogus",
		"menuFile":            "Bestand",
		"menuExit":            "Afsluiten",
		"menuSetup":           "Instellingen",
		"menuConfigUI":        "UI configureren",
		"menuHelp":            "Help",
		"menuAbout":           "Over",
		"language":            "Taal",
		"theme":               "Thema",
		"panelControls":       "Voorkeuren",
		"panelPreview":        "Hoofdgebied",
		"previewText":         "Visuele catalogus (binnenkort)\n\nGames, muziek, afbeeldingen en metadata komen hier.",
		"statusReady":         "Gereed.",
		"statusLogTitle":      "Status en debug",
		"configTitle":         "UI configureren",
		"configFontFamily":    "Lettertypefamilie",
		"configFontSize":      "Lettergrootte",
		"configFontSizeValue": "%d px",
		"configDensity":       "Lay-outdichtheid",
		"configOK":            "Toepassen",
		"configCancel":        "Annuleren",
		"aboutTitle":          "Over",
		"aboutVersion":        "Versie",
		"aboutBuild":          "Build",
		"aboutDate":           "Datum",
		"aboutTime":           "Tijd",
		"aboutCopyright1":     "(C) WIB Projetos Ltda.",
		"aboutCopyright2":     "(C) Cybernostra, Inc.",
		"aboutWebsite":        "www.cybernostra.com",
		"aboutYears":          "1972 - %d",
		"aboutClose":          "Sluiten",
		"cliShort":            "MSX database downloader en visuele catalogus",
		"cliLong":             "MSX DB Down downloadt MSX-softwaredatabases, verrijkt items met metadata (afbeeldingen, muziek, video, release-info) en biedt een visuele catalogus als frontend voor MSX-emulators.\n\nZonder subcommando wordt de grafische interface geopend.",
		"cliFlagLang":         "Stel UI-taal in: pt | en | es | nl | it",
		"cliFlagTheme":        "Stel UI-thema in: system | light | dark",
		"cliFlagDebug":        "Toon extra debugberichten in het logpaneel",
		"cliVersionShort":     "Toon versie-informatie",
	},
	uiprefs.LangIT: {
		"appTitle":            "MSX DB Down - Catalogo",
		"menuFile":            "File",
		"menuExit":            "Esci",
		"menuSetup":           "Impostazioni",
		"menuConfigUI":        "Configura UI",
		"menuHelp":            "Aiuto",
		"menuAbout":           "Informazioni",
		"language":            "Lingua",
		"theme":               "Tema",
		"panelControls":       "Preferenze",
		"panelPreview":        "Area principale",
		"previewText":         "Catalogo visivo (prossimamente)\n\nQui verranno mostrati giochi, musica, immagini e metadati.",
		"statusReady":         "Pronto.",
		"statusLogTitle":      "Stato e debug",
		"configTitle":         "Configura interfaccia",
		"configFontFamily":    "Famiglia di caratteri",
		"configFontSize":      "Dimensione carattere",
		"configFontSizeValue": "%d px",
		"configDensity":       "Densità del layout",
		"configOK":            "Applica",
		"configCancel":        "Annulla",
		"aboutTitle":          "Informazioni",
		"aboutVersion":        "Versione",
		"aboutBuild":          "Build",
		"aboutDate":           "Data",
		"aboutTime":           "Ora",
		"aboutCopyright1":     "(C) WIB Projetos Ltda.",
		"aboutCopyright2":     "(C) Cybernostra, Inc.",
		"aboutWebsite":        "www.cybernostra.com",
		"aboutYears":          "1972 - %d",
		"aboutClose":          "Chiudi",
		"cliShort":            "Downloader database MSX e catalogo visuale",
		"cliLong":             "MSX DB Down scarica database software MSX, arricchisce gli elementi con metadati (immagini, musica, video, info di rilascio) e fornisce un catalogo visuale come frontend per emulatori MSX.\n\nEseguire senza sottocomando apre l'interfaccia grafica.",
		"cliFlagLang":         "Imposta lingua UI: pt | en | es | nl | it",
		"cliFlagTheme":        "Imposta tema UI: system | light | dark",
		"cliFlagDebug":        "Mostra messaggi di debug aggiuntivi nel pannello log",
		"cliVersionShort":     "Mostra informazioni versione",
	},
}

type mainUI struct {
	app    fyne.App
	window fyne.Window
	store  *settingsdb.Store

	currentLanguage uiprefs.LanguageCode

	statusLine *widget.Label
	logEntry   *widget.Entry

	languageLabel *widget.Label
	themeLabel    *widget.Label
	controlsCard  *widget.Card
	previewCard   *widget.Card
	logCard       *widget.Card
	previewText   *widget.RichText
}

func main() {
	cliLang := detectCLILanguage(os.Args[1:])
	rootCmd := buildRootCmd(cliLang)
	rootCmd.AddCommand(buildVersionCmd(cliLang))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// buildRootCmd creates the root Cobra command, which launches the GUI.
func buildRootCmd(cliLang uiprefs.LanguageCode) *cobra.Command {
	tcli := func(key string) string { return tForLanguage(cliLang, key) }

	cmd := &cobra.Command{
		Use:          "msxdbdown",
		Short:        tcli("cliShort"),
		Long:         tcli("cliLong"),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			langFlag, _ := cmd.Flags().GetString("lang")
			themeFlag, _ := cmd.Flags().GetString("theme")
			debugFlag, _ := cmd.Flags().GetBool("debug")

			store, err := settingsdb.OpenDefault()
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()

			application := app.NewWithID("com.msxdbdown.gui")

			// CLI flags override stored settings and are persisted in SQLite.
			if langFlag != "" {
				if lang, ok := parseCLILanguage(langFlag); ok {
					_ = store.Set(prefLanguage, string(lang))
				} else {
					fmt.Fprintf(os.Stderr, "warning: unknown language %q — ignored\n", langFlag)
				}
			}
			if themeFlag != "" {
				normalized := uiprefs.ReadTheme(themeFlag)
				if strings.ToLower(themeFlag) == "light" || strings.ToLower(themeFlag) == "dark" || strings.ToLower(themeFlag) == "system" {
					_ = store.Set(prefTheme, normalized)
				} else {
					fmt.Fprintf(os.Stderr, "warning: unknown theme %q — ignored (use: system|light|dark)\n", themeFlag)
				}
			}

			ui := newMainUI(application, store)

			if debugFlag {
				ui.appendLog("CLI", fmt.Sprintf("msxdbdown %s build %s", AppVersion, BuildNumber))
				ui.appendLog("CLI", "Debug mode enabled via --debug flag")
			}

			ui.window.ShowAndRun()
			return nil
		},
	}

	cmd.Flags().StringP("lang", "l", "", tcli("cliFlagLang"))
	cmd.Flags().StringP("theme", "t", "", tcli("cliFlagTheme"))
	cmd.Flags().BoolP("debug", "d", false, tcli("cliFlagDebug"))

	return cmd
}

// buildVersionCmd creates the "version" sub-command.
func buildVersionCmd(cliLang uiprefs.LanguageCode) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: tForLanguage(cliLang, "cliVersionShort"),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("msxdbdown  version : %s\n", AppVersion)
			fmt.Printf("           built   : %s\n", BuildDate)
			fmt.Printf("           build # : %s\n", BuildNumber)
		},
	}
}

func tForLanguage(lang uiprefs.LanguageCode, key string) string {
	if tr, ok := i18n[lang][key]; ok {
		return tr
	}
	if fallback, ok := i18n[uiprefs.LangEN][key]; ok {
		return fallback
	}
	return key
}

func detectCLILanguage(args []string) uiprefs.LanguageCode {
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		switch {
		case arg == "--lang" || arg == "-l":
			if i+1 < len(args) {
				if lang, ok := parseCLILanguage(args[i+1]); ok {
					return lang
				}
			}
		case strings.HasPrefix(arg, "--lang="):
			if lang, ok := parseCLILanguage(strings.TrimPrefix(arg, "--lang=")); ok {
				return lang
			}
		case strings.HasPrefix(arg, "-l="):
			if lang, ok := parseCLILanguage(strings.TrimPrefix(arg, "-l=")); ok {
				return lang
			}
		}
	}

	return uiprefs.LangEN
}

func parseCLILanguage(value string) (uiprefs.LanguageCode, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(uiprefs.LangPT):
		return uiprefs.LangPT, true
	case string(uiprefs.LangEN):
		return uiprefs.LangEN, true
	case string(uiprefs.LangES):
		return uiprefs.LangES, true
	case string(uiprefs.LangNL):
		return uiprefs.LangNL, true
	case string(uiprefs.LangIT):
		return uiprefs.LangIT, true
	default:
		return "", false
	}
}

func newMainUI(a fyne.App, store *settingsdb.Store) *mainUI {
	ui := &mainUI{
		app:    a,
		window: a.NewWindow(""),
		store:  store,
	}

	ui.currentLanguage = uiprefs.ReadLanguage(ui.getSetting(prefLanguage))
	ui.window.Resize(fyne.NewSize(1180, 760))
	ui.window.CenterOnScreen()

	ui.logEntry = widget.NewMultiLineEntry()
	ui.logEntry.Wrapping = fyne.TextWrapWord
	ui.logEntry.Disable()

	ui.statusLine = widget.NewLabel("")
	ui.languageLabel = widget.NewLabel("")
	ui.themeLabel = widget.NewLabel("")
	initializing := true

	langSelect := widget.NewSelect(uiprefs.LanguageOptions(), func(selected string) {
		if initializing {
			return
		}
		lang := uiprefs.DisplayToLanguage[selected]
		ui.currentLanguage = lang
		ui.setSetting(prefLanguage, string(lang))
		ui.applyLanguage()
		ui.appendLog("UI", fmt.Sprintf("Language changed to %s", string(lang)))
	})

	themeSelect := widget.NewSelect([]string{"System", "Light", "Dark"}, func(selected string) {
		if initializing {
			return
		}
		ui.setSetting(prefTheme, selected)
		ui.applyTheme(selected)
		ui.appendLog("UI", fmt.Sprintf("Theme changed to %s", selected))
	})

	controlsContent := container.NewVBox(
		ui.languageLabel,
		langSelect,
		widget.NewSeparator(),
		ui.themeLabel,
		themeSelect,
	)
	ui.controlsCard = widget.NewCard("", "", controlsContent)

	ui.previewText = widget.NewRichTextFromMarkdown("")
	ui.previewText.Wrapping = fyne.TextWrapWord
	ui.previewCard = widget.NewCard("", "", container.NewPadded(ui.previewText))

	ui.logCard = widget.NewCard("", "", container.NewVScroll(ui.logEntry))

	mainSplit := container.NewHSplit(ui.controlsCard, container.NewVSplit(ui.previewCard, ui.logCard))
	mainSplit.Offset = 0.28

	root := container.NewBorder(nil, widget.NewCard("", "", ui.statusLine), nil, nil, mainSplit)
	ui.window.SetContent(root)
	langSelect.SetSelected(uiprefs.LanguageDisplay[ui.currentLanguage])
	themeSelect.SetSelected(uiprefs.ReadTheme(ui.getSetting(prefTheme)))
	initializing = false

	ui.applyAll()
	ui.applyLanguage()
	ui.window.SetMainMenu(ui.buildMenu())

	ui.appendLog("APP", "Main window initialized")
	return ui
}

func (ui *mainUI) applyLanguage() {
	tr := ui.t
	ui.window.SetTitle(tr("appTitle"))
	ui.languageLabel.SetText(tr("language"))
	ui.themeLabel.SetText(tr("theme"))
	ui.controlsCard.Title = tr("panelControls")
	ui.previewCard.Title = tr("panelPreview")
	ui.logCard.Title = tr("statusLogTitle")

	ui.statusLine.SetText(tr("statusReady"))

	ui.previewText.ParseMarkdown(tr("previewText"))

	ui.controlsCard.Refresh()
	ui.previewCard.Refresh()
	ui.logCard.Refresh()
	ui.window.SetMainMenu(ui.buildMenu())
}

func (ui *mainUI) applyTheme(themeName string) {
	ui.applyAll()
}

// applyAll builds a CustomTheme from the currently stored preferences and
// applies it to the running application. Call this whenever theme, font size,
// density or font family changes.
func (ui *mainUI) applyAll() {
	var base fyne.Theme
	switch strings.ToLower(uiprefs.ReadTheme(ui.getSetting(prefTheme))) {
	case "light":
		base = theme.LightTheme()
	case "dark":
		base = theme.DarkTheme()
	default:
		base = theme.DefaultTheme()
	}
	fontSize := uiprefs.ReadFontSize(ui.getSetting(uiprefs.PrefFontSize))
	density := uiprefs.ReadDensity(ui.getSetting(uiprefs.PrefDensity))
	fontName := uiprefs.ReadFontName(ui.getSetting(uiprefs.PrefFontName))
	ui.app.Settings().SetTheme(uitheme.New(base, fontSize, density, fontName))
}

func (ui *mainUI) buildMenu() *fyne.MainMenu {
	tr := ui.t

	fileMenu := fyne.NewMenu(tr("menuFile"),
		fyne.NewMenuItem(tr("menuExit"), func() {
			ui.appendLog("APP", "Exit requested")
			ui.app.Quit()
		}),
	)

	setupMenu := fyne.NewMenu(tr("menuSetup"),
		fyne.NewMenuItem(tr("menuConfigUI"), func() {
			configui.Show(ui.window, ui.getSetting, ui.setSetting, ui.t, func() {
				ui.applyAll()
				ui.appendLog("UI", "Config UI applied")
			})
		}),
	)

	helpMenu := fyne.NewMenu(tr("menuHelp"),
		fyne.NewMenuItem(tr("menuAbout"), func() {
			currentYear := time.Now().Year()
			yearStr := fmt.Sprintf(tr("aboutYears"), currentYear)

			// Wrap tr so aboutYears already has the year substituted
			trAbout := func(key string) string {
				if key == "aboutYears" {
					return yearStr
				}
				return tr(key)
			}

			about.Show(ui.window, AppVersion, BuildDate, BuildTime, BuildNumber, trAbout)
		}),
	)

	return fyne.NewMainMenu(fileMenu, setupMenu, helpMenu)
}

func (ui *mainUI) appendLog(scope string, message string) {
	timestamp := time.Now().Format("15:04:05")
	entry := fmt.Sprintf("[%s] [%s] %s", timestamp, scope, message)

	current := strings.TrimSpace(ui.logEntry.Text)
	if current == "" {
		ui.logEntry.SetText(entry)
	} else {
		ui.logEntry.SetText(current + "\n" + entry)
	}

	ui.statusLine.SetText(message)
}

func (ui *mainUI) t(key string) string {
	if tr, ok := i18n[ui.currentLanguage][key]; ok {
		return tr
	}
	if fallback, ok := i18n[uiprefs.LangEN][key]; ok {
		return fallback
	}
	return key
}

func (ui *mainUI) getSetting(key string) string {
	if ui.store == nil {
		return ""
	}
	value, err := ui.store.Get(key)
	if err != nil {
		return ""
	}
	return value
}

func (ui *mainUI) setSetting(key, value string) {
	if ui.store == nil {
		return
	}
	_ = ui.store.Set(key, value)
}
