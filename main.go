package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

	viewCatalogPlaceholder = "catalog"
	viewMSXRomDBUpdate     = "db.msxromdb"
	viewFileHunterUpdate   = "db.filehunter"
	viewCleanDownloads     = "db.clean"

	statusInfo  = "INFO"
	statusWarn  = "WARN"
	statusError = "ERROR"
)

var i18n = map[uiprefs.LanguageCode]map[string]string{
	uiprefs.LangPT: {
		"appTitle":                "MSX DB Down - Catálogo",
		"menuFile":                "Arquivo",
		"menuExit":                "Sair",
		"menuSetup":               "Configuração",
		"menuConfigUI":            "Configurar UI",
		"menuDatabase":            "Banco de Dados",
		"menuUpdateMSXRomDB":      "Atualizar MSX RomDB",
		"menuUpdateFileHunter":    "Atualizar File-Hunter",
		"menuHelp":                "Ajuda",
		"menuAbout":               "Sobre",
		"language":                "Idioma",
		"theme":                   "Tema",
		"panelControls":           "Preferências",
		"panelPreview":            "Área principal",
		"previewText":             "Catálogo visual (em breve)\n\nAqui serão exibidos jogos, músicas, imagens e metadados.",
		"statusReady":             "Pronto.",
		"statusLogTitle":          "Status e depuração",
		"configTitle":             "Configurar Interface",
		"configFontFamily":        "Família de Fonte",
		"configFontSize":          "Tamanho de Fonte",
		"configFontSizeValue":     "%d px",
		"configDensity":           "Densidade de Layout",
		"configMSXRomDBURL":       "MSX ROM DB URL",
		"configFileHunterURL":     "File-Hunter URL",
		"configFileHunterSHAURL":  "File-Hunter SHA URL",
		"configTabUI":             "UI",
		"configTabURLs":           "URLs",
		"configTabSQLite":         "SQLite",
		"configCurrentDBPath":     "Banco atual",
		"configCatalogDBLocation": "Local do banco de configuracoes",
		"configCatalogDBPath":     "Caminho do arquivo",
		"configDBLocationLocal":   "Diretorio local (data/msxdbdown.db)",
		"configDBLocationAppData": "Pasta de configuracao do usuario (%APPDATA% no Windows, ~/.config no Linux)",
		"configCreateCatalogDB":   "Criar banco inicial",
		"configDBPathError":       "Falha ao resolver caminho",
		"configDBSwitchTitle":     "Alterar local do banco",
		"configDBSwitchAskMove":   "Banco atual:\n%s\n\nNovo local:\n%s\n\nDeseja mover o banco atual para o novo local?",
		"configDBSwitchAskNew":    "Deseja criar um banco novo e zerado em:\n%s",
		"configDBExistsTitle":     "Banco ja existe",
		"configDBExistsConfirm":   "O banco ja existe em:\n%s\n\nDeseja recriar um banco zerado?",
		"configDBCreatedTitle":    "Banco inicial criado",
		"configDBCreatedMessage":  "Banco SQLite pronto em:\n%s",
		"configOK":                "Aplicar",
		"configCancel":            "Cancelar",
		"dbSourceURL":             "Endereço",
		"dbSourceSHAURL":          "SHA URL",
		"dbUpdateButton":          "Atualizar",
		"dbDownloadStarted":       "Iniciando download para download/ ...",
		"dbDownloadDone":          "Download concluído",
		"dbDownloadFailed":        "Falha no download",
		"dbExtractStarted":        "Descompactando arquivo em download/ ...",
		"dbExtractDone":           "Descompactação concluída",
		"dbExtractFailed":         "Falha na descompactação",
		"dbImportStarted":         "Importando SQL para o banco SQLite atual ...",
		"dbImportDone":            "Importação concluída",
		"dbImportFailed":          "Falha na importação SQL",
		"dbImportSQLNotFound":     "Arquivo SQL não encontrado após descompactação",
		"menuCleanDownloads":      "Limpar Downloads",
		"cleanDownloadsTitle":     "Limpar Downloads",
		"cleanDownloadsLabel":     "Arquivos em download/:",
		"cleanDownloadsButton":    "Limpar",
		"cleanDownloadsNoFiles":   "Nenhum arquivo em download/",
		"cleanDownloadsDone":      "Arquivos deletados com sucesso",
		"cleanDownloadsFailed":    "Erro ao deletar arquivos",
		"aboutTitle":              "Sobre",
		"aboutVersion":            "Versão",
		"aboutBuild":              "Build",
		"aboutDate":               "Data",
		"aboutTime":               "Hora",
		"aboutCopyright1":         "(C) WIB Projetos Ltda.",
		"aboutCopyright2":         "(C) Cybernostra, Inc.",
		"aboutWebsite":            "www.cybernostra.com",
		"aboutYears":              "1972 - %d",
		"aboutClose":              "Fechar",
		"cliShort":                "Baixador de banco de dados MSX e catalogo visual",
		"cliLong":                 "MSX DB Down baixa bancos de software MSX, enriquece itens com metadados (imagens, musica, video, informacoes de lancamento) e fornece um catalogo visual como frontend para emuladores MSX.\n\nExecutar sem subcomando abre a interface grafica.",
		"cliFlagLang":             "Define idioma da UI: pt | en | es | nl | it",
		"cliFlagTheme":            "Define tema da UI: system | light | dark",
		"cliFlagDebug":            "Mostra mensagens extras de depuracao no painel de log",
		"cliVersionShort":         "Mostra informacoes de versao",
	},
	uiprefs.LangEN: {
		"appTitle":                "MSX DB Down - Catalog",
		"menuFile":                "File",
		"menuExit":                "Exit",
		"menuSetup":               "Setup",
		"menuConfigUI":            "Config UI",
		"menuDatabase":            "Database",
		"menuUpdateMSXRomDB":      "Update MSX RomDB",
		"menuUpdateFileHunter":    "Update File-Hunter",
		"menuHelp":                "Help",
		"menuAbout":               "About",
		"language":                "Language",
		"theme":                   "Theme",
		"panelControls":           "Preferences",
		"panelPreview":            "Main area",
		"previewText":             "Visual catalog (coming soon)\n\nGames, music, images and metadata will be shown here.",
		"statusReady":             "Ready.",
		"statusLogTitle":          "Status and debug",
		"configTitle":             "Config UI",
		"configFontFamily":        "Font Family",
		"configFontSize":          "Font Size",
		"configFontSizeValue":     "%d px",
		"configDensity":           "Layout Density",
		"configMSXRomDBURL":       "MSX ROM DB URL",
		"configFileHunterURL":     "File-Hunter URL",
		"configFileHunterSHAURL":  "File-Hunter SHA URL",
		"configTabUI":             "UI",
		"configTabURLs":           "URLs",
		"configTabSQLite":         "SQLite",
		"configCurrentDBPath":     "Current database",
		"configCatalogDBLocation": "Settings database location",
		"configCatalogDBPath":     "Database file path",
		"configDBLocationLocal":   "Local directory (data/msxdbdown.db)",
		"configDBLocationAppData": "User config directory (%APPDATA% on Windows, ~/.config on Linux)",
		"configCreateCatalogDB":   "Create initial database",
		"configDBPathError":       "Could not resolve path",
		"configDBSwitchTitle":     "Switch database location",
		"configDBSwitchAskMove":   "Current database:\n%s\n\nTarget location:\n%s\n\nDo you want to move the current database?",
		"configDBSwitchAskNew":    "Do you want to create a new empty database at:\n%s",
		"configDBExistsTitle":     "Database already exists",
		"configDBExistsConfirm":   "Database already exists at:\n%s\n\nDo you want to recreate an empty database?",
		"configDBCreatedTitle":    "Initial database created",
		"configDBCreatedMessage":  "SQLite database ready at:\n%s",
		"configOK":                "Apply",
		"configCancel":            "Cancel",
		"dbSourceURL":             "Address",
		"dbSourceSHAURL":          "SHA URL",
		"dbUpdateButton":          "Update",
		"dbDownloadStarted":       "Starting download to download/ ...",
		"dbDownloadDone":          "Download completed",
		"dbDownloadFailed":        "Download failed",
		"dbExtractStarted":        "Extracting file into download/ ...",
		"dbExtractDone":           "Extraction completed",
		"dbExtractFailed":         "Extraction failed",
		"dbImportStarted":         "Importing SQL into the current SQLite database ...",
		"dbImportDone":            "Import finished",
		"dbImportFailed":          "SQL import failed",
		"dbImportSQLNotFound":     "SQL file not found after extraction",
		"menuCleanDownloads":      "Clean Downloads",
		"cleanDownloadsTitle":     "Clean Downloads",
		"cleanDownloadsLabel":     "Files in download/:",
		"cleanDownloadsButton":    "Clean",
		"cleanDownloadsNoFiles":   "No files in download/",
		"cleanDownloadsDone":      "Files deleted successfully",
		"cleanDownloadsFailed":    "Error deleting files",
		"aboutTitle":              "About",
		"aboutVersion":            "Version",
		"aboutBuild":              "Build",
		"aboutDate":               "Date",
		"aboutTime":               "Time",
		"aboutCopyright1":         "(C) WIB Projetos Ltda.",
		"aboutCopyright2":         "(C) Cybernostra, Inc.",
		"aboutWebsite":            "www.cybernostra.com",
		"aboutYears":              "1972 - %d",
		"aboutClose":              "Close",
		"cliShort":                "MSX Database Downloader and Visual Catalog",
		"cliLong":                 "MSX DB Down downloads MSX software databases, enriches entries with metadata (images, music, video, release info) and provides a visual catalog that acts as a frontend for MSX emulators.\n\nRunning without a sub-command opens the graphical interface.",
		"cliFlagLang":             "Override UI language: pt | en | es | nl | it",
		"cliFlagTheme":            "Override UI theme: system | light | dark",
		"cliFlagDebug":            "Print extra debug messages in the log panel",
		"cliVersionShort":         "Print version information",
	},
	uiprefs.LangES: {
		"appTitle":               "MSX DB Down - Catálogo",
		"menuFile":               "Archivo",
		"menuExit":               "Salir",
		"menuSetup":              "Configuración",
		"menuConfigUI":           "Configurar UI",
		"menuDatabase":           "Base de Datos",
		"menuUpdateMSXRomDB":     "Actualizar MSX RomDB",
		"menuUpdateFileHunter":   "Actualizar File-Hunter",
		"menuCleanDownloads":     "Limpiar Descargas",
		"menuHelp":               "Ayuda",
		"menuAbout":              "Acerca de",
		"language":               "Idioma",
		"theme":                  "Tema",
		"panelControls":          "Preferencias",
		"panelPreview":           "Área principal",
		"previewText":            "Catálogo visual (próximamente)\n\nAquí se mostrarán juegos, música, imágenes y metadatos.",
		"statusReady":            "Listo.",
		"statusLogTitle":         "Estado y depuración",
		"configTitle":            "Configurar interfaz",
		"configFontFamily":       "Familia de fuente",
		"configFontSize":         "Tamaño de fuente",
		"configFontSizeValue":    "%d px",
		"configDensity":          "Densidad de diseño",
		"configMSXRomDBURL":      "URL de MSX ROM DB",
		"configFileHunterURL":    "URL de File-Hunter",
		"configFileHunterSHAURL": "URL de SHA de File-Hunter",
		"configOK":               "Aplicar",
		"configCancel":           "Cancelar",
		"dbSourceURL":            "Dirección",
		"dbSourceSHAURL":         "URL SHA",
		"dbUpdateButton":         "Actualizar",
		"dbDownloadStarted":      "Iniciando descarga a descargas/ ...",
		"dbDownloadDone":         "Descarga completada",
		"dbDownloadFailed":       "Falha en la descarga",
		"dbExtractStarted":       "Extrayendo archivo en download/ ...",
		"dbExtractDone":          "Extracción completada",
		"dbExtractFailed":        "Fallo al extraer archivo",
		"cleanDownloadsTitle":    "Limpiar Descargas",
		"cleanDownloadsLabel":    "Archivos en descargas/:",
		"cleanDownloadsButton":   "Limpiar",
		"cleanDownloadsNoFiles":  "Sin archivos en descargas/",
		"cleanDownloadsDone":     "Archivos eliminados exitosamente",
		"cleanDownloadsFailed":   "Error al eliminar archivos",
		"aboutTitle":             "Acerca de",
		"aboutVersion":           "Versión",
		"aboutBuild":             "Build",
		"aboutDate":              "Fecha",
		"aboutTime":              "Hora",
		"aboutCopyright1":        "(C) WIB Projetos Ltda.",
		"aboutCopyright2":        "(C) Cybernostra, Inc.",
		"aboutWebsite":           "www.cybernostra.com",
		"aboutYears":             "1972 - %d",
		"aboutClose":             "Cerrar",
		"cliShort":               "Descargador de base de datos MSX y catálogo visual",
		"cliLong":                "MSX DB Down descarga bases de software MSX, enriquece elementos con metadatos (imágenes, música, video, información de lanzamiento) y ofrece un catálogo visual como frontend para emuladores MSX.\n\nEjecutar sin subcomando abre la interfaz gráfica.",
		"cliFlagLang":            "Define idioma de la UI: pt | en | es | nl | it",
		"cliFlagTheme":           "Define tema de la UI: system | light | dark",
		"cliFlagDebug":           "Muestra mensajes extra de depuración en el panel de registro",
		"cliVersionShort":        "Muestra información de versión",
	},
	uiprefs.LangNL: {
		"appTitle":               "MSX DB Down - Catalogus",
		"menuFile":               "Bestand",
		"menuExit":               "Afsluiten",
		"menuSetup":              "Instellingen",
		"menuConfigUI":           "UI configureren",
		"menuDatabase":           "Gegevensbank",
		"menuUpdateMSXRomDB":     "MSX RomDB bijwerken",
		"menuUpdateFileHunter":   "File-Hunter bijwerken",
		"menuCleanDownloads":     "Downloads wissen",
		"menuHelp":               "Help",
		"menuAbout":              "Over",
		"language":               "Taal",
		"theme":                  "Thema",
		"panelControls":          "Voorkeuren",
		"panelPreview":           "Hoofdgebied",
		"previewText":            "Visuele catalogus (binnenkort)\n\nGames, muziek, afbeeldingen en metadata komen hier.",
		"statusReady":            "Gereed.",
		"statusLogTitle":         "Status en debug",
		"configTitle":            "UI configureren",
		"configFontFamily":       "Lettertypefamilie",
		"configFontSize":         "Lettergrootte",
		"configFontSizeValue":    "%d px",
		"configDensity":          "Lay-outdichtheid",
		"configMSXRomDBURL":      "MSX ROM DB URL",
		"configFileHunterURL":    "File-Hunter URL",
		"configFileHunterSHAURL": "File-Hunter SHA URL",
		"configOK":               "Toepassen",
		"configCancel":           "Annuleren",
		"dbSourceURL":            "Adres",
		"dbSourceSHAURL":         "SHA URL",
		"dbUpdateButton":         "Bijwerken",
		"dbDownloadStarted":      "Download starten naar downloads/ ...",
		"dbDownloadDone":         "Download voltooid",
		"dbDownloadFailed":       "Download mislukt",
		"dbExtractStarted":       "Bestand uitpakken naar download/ ...",
		"dbExtractDone":          "Uitpakken voltooid",
		"dbExtractFailed":        "Uitpakken mislukt",
		"cleanDownloadsTitle":    "Downloads wissen",
		"cleanDownloadsLabel":    "Bestanden in downloads/:",
		"cleanDownloadsButton":   "Wissen",
		"cleanDownloadsNoFiles":  "Geen bestanden in downloads/",
		"cleanDownloadsDone":     "Bestanden succesvol verwijderd",
		"cleanDownloadsFailed":   "Fout bij verwijderen van bestanden",
		"aboutTitle":             "Over",
		"aboutVersion":           "Versie",
		"aboutBuild":             "Build",
		"aboutDate":              "Datum",
		"aboutTime":              "Tijd",
		"aboutCopyright1":        "(C) WIB Projetos Ltda.",
		"aboutCopyright2":        "(C) Cybernostra, Inc.",
		"aboutWebsite":           "www.cybernostra.com",
		"aboutYears":             "1972 - %d",
		"aboutClose":             "Sluiten",
		"cliShort":               "MSX database downloader en visuele catalogus",
		"cliLong":                "MSX DB Down downloadt MSX-softwaredatabases, verrijkt items met metadata (afbeeldingen, muziek, video, release-info) en biedt een visuele catalogus als frontend voor MSX-emulators.\n\nZonder subcommando wordt de grafische interface geopend.",
		"cliFlagLang":            "Stel UI-taal in: pt | en | es | nl | it",
		"cliFlagTheme":           "Stel UI-thema in: system | light | dark",
		"cliFlagDebug":           "Toon extra debugberichten in het logpaneel",
		"cliVersionShort":        "Toon versie-informatie",
	},
	uiprefs.LangIT: {
		"appTitle":               "MSX DB Down - Catalogo",
		"menuFile":               "File",
		"menuExit":               "Esci",
		"menuSetup":              "Impostazioni",
		"menuConfigUI":           "Configura UI",
		"menuDatabase":           "Database",
		"menuUpdateMSXRomDB":     "Aggiorna MSX RomDB",
		"menuUpdateFileHunter":   "Aggiorna File-Hunter",
		"menuCleanDownloads":     "Pulisci Download",
		"menuHelp":               "Aiuto",
		"menuAbout":              "Informazioni",
		"language":               "Lingua",
		"theme":                  "Tema",
		"panelControls":          "Preferenze",
		"panelPreview":           "Area principale",
		"previewText":            "Catalogo visivo (prossimamente)\n\nQui verranno mostrati giochi, musica, immagini e metadati.",
		"statusReady":            "Pronto.",
		"statusLogTitle":         "Stato e debug",
		"configTitle":            "Configura interfaccia",
		"configFontFamily":       "Famiglia di caratteri",
		"configFontSize":         "Dimensione carattere",
		"configFontSizeValue":    "%d px",
		"configDensity":          "Densità del layout",
		"configMSXRomDBURL":      "URL del database MSX ROM",
		"configFileHunterURL":    "URL di File-Hunter",
		"configFileHunterSHAURL": "URL SHA di File-Hunter",
		"configOK":               "Applica",
		"configCancel":           "Annulla",
		"dbSourceURL":            "Indirizzo",
		"dbSourceSHAURL":         "URL SHA",
		"dbUpdateButton":         "Aggiorna",
		"dbDownloadStarted":      "Avvio download in download/ ...",
		"dbDownloadDone":         "Download completato",
		"dbDownloadFailed":       "Download non riuscito",
		"dbExtractStarted":       "Estrazione file in download/ ...",
		"dbExtractDone":          "Estrazione completata",
		"dbExtractFailed":        "Estrazione non riuscita",
		"cleanDownloadsTitle":    "Pulisci Download",
		"cleanDownloadsLabel":    "File in download/:",
		"cleanDownloadsButton":   "Pulisci",
		"cleanDownloadsNoFiles":  "Nessun file in download/",
		"cleanDownloadsDone":     "File eliminati con successo",
		"cleanDownloadsFailed":   "Errore eliminazione file",
		"aboutTitle":             "Informazioni",
		"aboutVersion":           "Versione",
		"aboutBuild":             "Build",
		"aboutDate":              "Data",
		"aboutTime":              "Ora",
		"aboutCopyright1":        "(C) WIB Projetos Ltda.",
		"aboutCopyright2":        "(C) Cybernostra, Inc.",
		"aboutWebsite":           "www.cybernostra.com",
		"aboutYears":             "1972 - %d",
		"aboutClose":             "Chiudi",
		"cliShort":               "Downloader database MSX e catalogo visuale",
		"cliLong":                "MSX DB Down scarica database software MSX, arricchisce gli elementi con metadati (immagini, musica, video, info di rilascio) e fornisce un catalogo visuale come frontend per emulatori MSX.\n\nEseguire senza sottocomando apre l'interfaccia grafica.",
		"cliFlagLang":            "Imposta lingua UI: pt | en | es | nl | it",
		"cliFlagTheme":           "Imposta tema UI: system | light | dark",
		"cliFlagDebug":           "Mostra messaggi di debug aggiuntivi nel pannello log",
		"cliVersionShort":        "Mostra informazioni versione",
	},
}

type mainUI struct {
	app        fyne.App
	window     fyne.Window
	store      *settingsdb.Store
	dbPath     string
	dbLocation string

	currentLanguage uiprefs.LanguageCode

	statusLine *widget.Label
	statusBG   *canvas.Rectangle
	logBG      *canvas.Rectangle
	logEntry   *widget.Entry
	logLines   []string

	languageLabel *widget.Label
	themeLabel    *widget.Label
	controlsCard  *widget.Card
	previewCard   *widget.Card
	logCard       *widget.Card
	previewText   *widget.RichText
	activeView    string
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

			dbLocation, dbPath, err := settingsdb.DetectCurrentPath()
			if err != nil {
				return err
			}

			store, err := settingsdb.Open(dbPath)
			if err != nil {
				return err
			}
			_ = store.Set(uiprefs.PrefCatalogDBLocation, dbLocation)

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

			ui := newMainUI(application, store, dbLocation, dbPath)

			if debugFlag {
				ui.appendLog(statusInfo, "CLI", fmt.Sprintf("msxdbdown %s build %s", AppVersion, BuildNumber))
				ui.appendLog(statusInfo, "CLI", "Debug mode enabled via --debug flag")
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

func newMainUI(a fyne.App, store *settingsdb.Store, dbLocation, dbPath string) *mainUI {
	ui := &mainUI{
		app:        a,
		window:     a.NewWindow(""),
		store:      store,
		dbPath:     dbPath,
		dbLocation: settingsdb.NormalizeLocation(dbLocation),
		activeView: viewCatalogPlaceholder,
	}
	ui.window.SetOnClosed(func() {
		if ui.store != nil {
			_ = ui.store.Close()
		}
	})

	ui.currentLanguage = uiprefs.ReadLanguage(ui.getSetting(prefLanguage))
	ui.window.Resize(fyne.NewSize(1180, 760))
	ui.window.CenterOnScreen()

	ui.logEntry = widget.NewMultiLineEntry()
	ui.logEntry.Wrapping = fyne.TextWrapWord
	ui.logEntry.Disable()
	ui.logLines = []string{}

	ui.statusLine = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
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
		ui.appendLog(statusInfo, "UI", fmt.Sprintf("Language changed to %s", string(lang)))
	})

	themeSelect := widget.NewSelect([]string{"System", "Light", "Dark"}, func(selected string) {
		if initializing {
			return
		}
		ui.setSetting(prefTheme, selected)
		ui.applyTheme(selected)
		ui.appendLog(statusInfo, "UI", fmt.Sprintf("Theme changed to %s", selected))
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

	logBG := canvas.NewRectangle(color.NRGBA{R: 0xF5, G: 0xF7, B: 0xFA, A: 0xFF})
	logContent := container.NewStack(
		logBG,
		container.NewPadded(container.NewVScroll(ui.logEntry)),
	)
	ui.logBG = logBG
	ui.logCard = widget.NewCard("", "", logContent)

	mainSplit := container.NewHSplit(ui.controlsCard, container.NewVSplit(ui.previewCard, ui.logCard))
	mainSplit.Offset = 0.28

	ui.statusBG = canvas.NewRectangle(statusColorFor(statusInfo))
	statusBar := container.NewStack(ui.statusBG, container.NewPadded(ui.statusLine))
	root := container.NewBorder(nil, statusBar, nil, nil, mainSplit)
	ui.window.SetContent(root)
	langSelect.SetSelected(uiprefs.LanguageDisplay[ui.currentLanguage])
	themeSelect.SetSelected(uiprefs.ReadTheme(ui.getSetting(prefTheme)))
	initializing = false

	ui.applyAll()
	ui.applyLanguage()
	ui.window.SetMainMenu(ui.buildMenu())

	ui.updateLogPanelContrast()
	ui.appendLog(statusInfo, "APP", "Main window initialized")
	return ui
}

func (ui *mainUI) applyLanguage() {
	tr := ui.t
	ui.window.SetTitle(tr("appTitle"))
	ui.languageLabel.SetText(tr("language"))
	ui.themeLabel.SetText(tr("theme"))
	ui.controlsCard.Title = tr("panelControls")
	ui.logCard.Title = tr("statusLogTitle")

	ui.setStatus(statusInfo, tr("statusReady"))
	ui.renderActiveView()

	ui.controlsCard.Refresh()
	ui.previewCard.Refresh()
	ui.logCard.Refresh()
	ui.window.SetMainMenu(ui.buildMenu())
}

func (ui *mainUI) applyTheme(themeName string) {
	ui.applyAll()
	ui.updateLogPanelContrast()
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

func (ui *mainUI) renderActiveView() {
	switch ui.activeView {
	case viewMSXRomDBUpdate:
		ui.showMSXRomDBUpdateView()
	case viewFileHunterUpdate:
		ui.showFileHunterUpdateView()
	case viewCleanDownloads:
		ui.showCleanDownloadsView()
	default:
		ui.activeView = viewCatalogPlaceholder
		ui.previewCard.Title = ui.t("panelPreview")
		ui.previewText.ParseMarkdown(ui.t("previewText"))
		ui.previewCard.SetContent(container.NewPadded(ui.previewText))
	}
}

func (ui *mainUI) showMSXRomDBUpdateView() {
	ui.activeView = viewMSXRomDBUpdate
	romURL := ui.resolveURLSetting(uiprefs.PrefMSXRomDBURL, uiprefs.DefaultMSXRomDBURL, "MSX RomDB")

	urlLabel := widget.NewLabel(romURL)
	urlLabel.Wrapping = fyne.TextWrapWord

	updateBtn := widget.NewButton(ui.t("dbUpdateButton"), func() {
		romURL = ui.resolveURLSetting(uiprefs.PrefMSXRomDBURL, uiprefs.DefaultMSXRomDBURL, "MSX RomDB")
		ui.appendLog(statusInfo, "DB", ui.t("dbDownloadStarted"))
		savedPath, err := ui.downloadSingleFile(romURL)
		if err != nil {
			ui.appendLog(statusError, "DB", fmt.Sprintf("%s: %v", ui.t("dbDownloadFailed"), err))
			dialog.ShowError(err, ui.window)
			return
		}
		ui.appendLog(statusInfo, "DB", fmt.Sprintf("%s: %s", ui.t("dbDownloadDone"), savedPath))

		ui.appendLog(statusInfo, "DB", ui.t("dbExtractStarted"))
		extractedPaths, err := extractZipFlat(savedPath, filepath.Join(".", "download"))
		if err != nil {
			ui.appendLog(statusError, "DB", fmt.Sprintf("%s: %v", ui.t("dbExtractFailed"), err))
			dialog.ShowError(err, ui.window)
			return
		}
		ui.appendLog(statusInfo, "DB", fmt.Sprintf("%s: %s", ui.t("dbExtractDone"), strings.Join(extractedPaths, ", ")))

		sqlDumpPath := ""
		for _, extractedPath := range extractedPaths {
			name := strings.ToLower(filepath.Base(extractedPath))
			if name == "sql-msxromdb.sql" || name == "sql-romdb.sql" {
				sqlDumpPath = extractedPath
				break
			}
			if sqlDumpPath == "" && strings.HasSuffix(name, ".sql") {
				sqlDumpPath = extractedPath
			}
		}
		if sqlDumpPath == "" {
			err := errors.New(ui.t("dbImportSQLNotFound"))
			ui.appendLog(statusError, "DB", err.Error())
			dialog.ShowError(err, ui.window)
			return
		}

		ui.appendLog(statusInfo, "DB", ui.t("dbImportStarted"))
		insertCount, err := ui.store.ImportSQLDump(sqlDumpPath, true, func(message string) {
			ui.appendLog(statusInfo, "DB", "Import: "+message)
		})
		if err != nil {
			ui.appendLog(statusError, "DB", fmt.Sprintf("%s: %v", ui.t("dbImportFailed"), err))
			dialog.ShowError(err, ui.window)
			return
		}
		ui.appendLog(statusInfo, "DB", fmt.Sprintf("%s (%d): %s", ui.t("dbImportDone"), insertCount, sqlDumpPath))
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle(ui.t("dbSourceURL"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		urlLabel,
		updateBtn,
	)

	ui.previewCard.Title = ui.t("menuUpdateMSXRomDB")
	ui.previewCard.SetContent(container.NewPadded(content))
	ui.previewCard.Refresh()
}

func (ui *mainUI) showFileHunterUpdateView() {
	ui.activeView = viewFileHunterUpdate
	fileHunterURL := ui.resolveURLSetting(uiprefs.PrefFileHunterURL, uiprefs.DefaultFileHunterURL, "File-Hunter")
	fileHunterSHAURL := ui.resolveURLSetting(uiprefs.PrefFileHunterSHAURL, uiprefs.DefaultFileHunterSHAURL, "File-Hunter SHA")

	urlLabel := widget.NewLabel(fileHunterURL)
	urlLabel.Wrapping = fyne.TextWrapWord

	shaLabel := widget.NewLabel(fileHunterSHAURL)
	shaLabel.Wrapping = fyne.TextWrapWord

	updateBtn := widget.NewButton(ui.t("dbUpdateButton"), func() {
		fileHunterURL = ui.resolveURLSetting(uiprefs.PrefFileHunterURL, uiprefs.DefaultFileHunterURL, "File-Hunter")
		fileHunterSHAURL = ui.resolveURLSetting(uiprefs.PrefFileHunterSHAURL, uiprefs.DefaultFileHunterSHAURL, "File-Hunter SHA")
		ui.appendLog(statusInfo, "DB", ui.t("dbDownloadStarted"))

		allFilesPath, err := ui.downloadSingleFile(fileHunterURL)
		if err != nil {
			ui.appendLog(statusError, "DB", fmt.Sprintf("%s: %v", ui.t("dbDownloadFailed"), err))
			dialog.ShowError(err, ui.window)
			return
		}

		shaPath, err := ui.downloadSingleFile(fileHunterSHAURL)
		if err != nil {
			ui.appendLog(statusError, "DB", fmt.Sprintf("%s: %v", ui.t("dbDownloadFailed"), err))
			dialog.ShowError(err, ui.window)
			return
		}

		ui.appendLog(statusInfo, "DB", fmt.Sprintf("%s: %s, %s", ui.t("dbDownloadDone"), allFilesPath, shaPath))
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle(ui.t("dbSourceURL"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		urlLabel,
		widget.NewSeparator(),
		widget.NewLabelWithStyle(ui.t("dbSourceSHAURL"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		shaLabel,
		updateBtn,
	)

	ui.previewCard.Title = ui.t("menuUpdateFileHunter")
	ui.previewCard.SetContent(container.NewPadded(content))
	ui.previewCard.Refresh()
}

func (ui *mainUI) showCleanDownloadsView() {
	ui.activeView = viewCleanDownloads

	downloadDir := filepath.Join(".", "download")
	files, err := os.ReadDir(downloadDir)
	if err != nil && !os.IsNotExist(err) {
		ui.appendLog(statusError, "DB", fmt.Sprintf("Error reading download dir: %v", err))
		return
	}

	if len(files) == 0 {
		noFilesLabel := widget.NewLabel(ui.t("cleanDownloadsNoFiles"))
		noFilesLabel.Alignment = fyne.TextAlignCenter
		ui.previewCard.Title = ui.t("cleanDownloadsTitle")
		ui.previewCard.SetContent(container.NewPadded(noFilesLabel))
		ui.previewCard.Refresh()
		return
	}

	var filesList []string
	for _, f := range files {
		if !f.IsDir() {
			filesList = append(filesList, f.Name())
		}
	}

	fileContent := container.NewVBox()
	fileContent.Add(widget.NewLabelWithStyle(ui.t("cleanDownloadsLabel"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

	for _, filename := range filesList {
		fileLabel := widget.NewLabel("  • " + filename)
		fileContent.Add(fileLabel)
	}

	cleanBtn := widget.NewButton(ui.t("cleanDownloadsButton"), func() {
		ui.appendLog(statusInfo, "DB", "Cleaning downloads...")
		downloadDir := filepath.Join(".", "download")
		entries, err := os.ReadDir(downloadDir)
		if err != nil {
			ui.appendLog(statusError, "DB", fmt.Sprintf("%s: %v", ui.t("cleanDownloadsFailed"), err))
			dialog.ShowError(err, ui.window)
			return
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				filePath := filepath.Join(downloadDir, entry.Name())
				if err := os.Remove(filePath); err != nil {
					ui.appendLog(statusError, "DB", fmt.Sprintf("%s: %v", ui.t("cleanDownloadsFailed"), err))
					dialog.ShowError(err, ui.window)
					return
				}
			}
		}

		ui.appendLog(statusInfo, "DB", ui.t("cleanDownloadsDone"))
		ui.showCleanDownloadsView()
	})

	content := container.NewVBox(
		container.NewVScroll(fileContent),
		widget.NewSeparator(),
		cleanBtn,
	)

	ui.previewCard.Title = ui.t("cleanDownloadsTitle")
	ui.previewCard.SetContent(container.NewPadded(content))
	ui.previewCard.Refresh()
}

func (ui *mainUI) downloadSingleFile(sourceURL string) (string, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(sourceURL))
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", fmt.Errorf("invalid URL: %q", sourceURL)
	}

	name := path.Base(parsedURL.Path)
	if name == "." || name == "/" || name == "" {
		name = "download.bin"
	}

	downloadDir := filepath.Join(".", "download")
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		return "", fmt.Errorf("create download dir: %w", err)
	}

	tempPath := filepath.Join(downloadDir, name+".part")
	finalPath := filepath.Join(downloadDir, name)

	const maxAttempts = 2
	client := &http.Client{Timeout: 180 * time.Second}
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			ui.appendLog(statusWarn, "DB", fmt.Sprintf("Retrying download (%d/%d): %s", attempt, maxAttempts, parsedURL.String()))
		}

		resp, reqErr := client.Get(parsedURL.String())
		if reqErr != nil {
			lastErr = fmt.Errorf("http get %s: %w", parsedURL.String(), reqErr)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			lastErr = fmt.Errorf("http get %s: status %s", parsedURL.String(), resp.Status)
			_ = resp.Body.Close()
			continue
		}

		f, createErr := os.Create(tempPath)
		if createErr != nil {
			_ = resp.Body.Close()
			return "", fmt.Errorf("create file %s: %w", tempPath, createErr)
		}

		_, copyErr := io.Copy(f, resp.Body)
		closeBodyErr := resp.Body.Close()
		closeFileErr := f.Close()
		if copyErr != nil {
			lastErr = fmt.Errorf("save file %s: %w", tempPath, copyErr)
			continue
		}
		if closeBodyErr != nil {
			lastErr = fmt.Errorf("close response body: %w", closeBodyErr)
			continue
		}
		if closeFileErr != nil {
			lastErr = fmt.Errorf("close file %s: %w", tempPath, closeFileErr)
			continue
		}

		if err := os.Rename(tempPath, finalPath); err != nil {
			lastErr = fmt.Errorf("finalize file %s: %w", finalPath, err)
			continue
		}

		return finalPath, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("download failed: %s", parsedURL.String())
	}
	return "", lastErr
}

func extractZipFlat(zipPath, destinationDir string) ([]string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("open zip %s: %w", zipPath, err)
	}
	defer func() { _ = reader.Close() }()

	if err := os.MkdirAll(destinationDir, 0o755); err != nil {
		return nil, fmt.Errorf("create extract dir %s: %w", destinationDir, err)
	}

	extracted := make([]string, 0, len(reader.File))
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		name := filepath.Base(file.Name)
		if name == "." || name == string(filepath.Separator) || name == "" {
			continue
		}

		source, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open zip entry %s: %w", file.Name, err)
		}

		tempPath := filepath.Join(destinationDir, name+".part")
		finalPath := filepath.Join(destinationDir, name)

		target, err := os.Create(tempPath)
		if err != nil {
			_ = source.Close()
			return nil, fmt.Errorf("create extracted file %s: %w", tempPath, err)
		}

		_, copyErr := io.Copy(target, source)
		closeSourceErr := source.Close()
		closeTargetErr := target.Close()
		if copyErr != nil {
			return nil, fmt.Errorf("extract file %s: %w", finalPath, copyErr)
		}
		if closeSourceErr != nil {
			return nil, fmt.Errorf("close zip entry %s: %w", file.Name, closeSourceErr)
		}
		if closeTargetErr != nil {
			return nil, fmt.Errorf("close extracted file %s: %w", finalPath, closeTargetErr)
		}

		if err := os.Rename(tempPath, finalPath); err != nil {
			return nil, fmt.Errorf("finalize extracted file %s: %w", finalPath, err)
		}

		extracted = append(extracted, finalPath)
	}

	if len(extracted) == 0 {
		return nil, fmt.Errorf("zip %s does not contain extractable files", zipPath)
	}

	return extracted, nil
}

func (ui *mainUI) resolveURLSetting(key, fallback, label string) string {
	stored := strings.TrimSpace(ui.getSetting(key))
	resolved := uiprefs.ReadURL(stored, fallback)
	if stored == "" {
		ui.appendLog(statusWarn, "DB", fmt.Sprintf("%s URL is empty; using default: %s", label, fallback))
	}
	return resolved
}

func (ui *mainUI) buildMenu() *fyne.MainMenu {
	tr := ui.t

	fileMenu := fyne.NewMenu(tr("menuFile"),
		fyne.NewMenuItem(tr("menuExit"), func() {
			ui.appendLog(statusInfo, "APP", "Exit requested")
			ui.app.Quit()
		}),
	)

	setupMenu := fyne.NewMenu(tr("menuSetup"),
		fyne.NewMenuItem(tr("menuConfigUI"), func() {
			configui.Show(
				ui.window,
				ui.getSetting,
				ui.setSetting,
				func() string { return ui.dbLocation },
				func() string { return ui.dbPath },
				settingsdb.ResolvePath,
				ui.switchSettingsDB,
				ui.t,
				func() {
					ui.applyAll()
					ui.renderActiveView()
					ui.appendLog(statusInfo, "UI", "Config UI applied")
				},
			)
		}),
	)

	databaseMenu := fyne.NewMenu(tr("menuDatabase"),
		fyne.NewMenuItem(tr("menuUpdateMSXRomDB"), func() {
			ui.showMSXRomDBUpdateView()
			ui.appendLog(statusInfo, "DB", "MSX RomDB update view opened")
		}),
		fyne.NewMenuItem(tr("menuUpdateFileHunter"), func() {
			ui.showFileHunterUpdateView()
			ui.appendLog(statusInfo, "DB", "File-Hunter update view opened")
		}),
		fyne.NewMenuItem(tr("menuCleanDownloads"), func() {
			ui.showCleanDownloadsView()
			ui.appendLog(statusInfo, "DB", "Clean downloads view opened")
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

	return fyne.NewMainMenu(fileMenu, setupMenu, databaseMenu, helpMenu)
}

func (ui *mainUI) appendLog(severity, scope, message string) {
	if severity == "" {
		severity = statusInfo
	}
	timestamp := time.Now().Format("15:04:05")
	entry := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, severity, scope, message)

	ui.logLines = append(ui.logLines, entry)

	content := strings.Join(ui.logLines, "\n")
	ui.logEntry.SetText(content)

	ui.setStatus(severity, message)
}

func (ui *mainUI) setStatus(severity, message string) {
	if severity == "" {
		severity = statusInfo
	}
	ui.statusLine.SetText(fmt.Sprintf("[%s] %s", severity, message))
	if ui.statusBG != nil {
		ui.statusBG.FillColor = statusColorFor(severity)
		ui.statusBG.Refresh()
	}
}

func (ui *mainUI) updateLogPanelContrast() {
	isDarkMode := strings.ToLower(uiprefs.ReadTheme(ui.getSetting(prefTheme))) == "dark"
	if isDarkMode {
		ui.logBG.FillColor = color.NRGBA{R: 0x2A, G: 0x30, B: 0x39, A: 0xFF}
	} else {
		ui.logBG.FillColor = color.NRGBA{R: 0xF5, G: 0xF7, B: 0xFA, A: 0xFF}
	}
	if ui.logBG != nil {
		ui.logBG.Refresh()
	}
	if ui.logEntry != nil {
		ui.logEntry.Refresh()
	}
}

func statusColorFor(severity string) color.NRGBA {
	switch strings.ToUpper(strings.TrimSpace(severity)) {
	case statusError:
		return color.NRGBA{R: 0xC6, G: 0x2D, B: 0x42, A: 0x66}
	case statusWarn:
		return color.NRGBA{R: 0xCF, G: 0x8A, B: 0x16, A: 0x66}
	default:
		return color.NRGBA{R: 0x3A, G: 0x78, B: 0xC2, A: 0x4D}
	}
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

func (ui *mainUI) switchSettingsDB(targetLocation string, moveCurrent bool) error {
	targetLocation = settingsdb.NormalizeLocation(targetLocation)
	targetPath, err := settingsdb.ResolvePath(targetLocation)
	if err != nil {
		return err
	}

	currentStore := ui.store
	currentPath := ui.dbPath
	currentLocation := ui.dbLocation

	if currentStore != nil {
		_ = currentStore.Close()
	}

	reopenCurrent := func(operationErr error) error {
		reopened, reopenErr := settingsdb.Open(currentPath)
		if reopenErr == nil {
			ui.store = reopened
			ui.dbPath = currentPath
			ui.dbLocation = currentLocation
			return operationErr
		}
		return fmt.Errorf("%w (also failed to reopen current db: %v)", operationErr, reopenErr)
	}

	if moveCurrent {
		if err := settingsdb.MoveDatabaseFiles(currentPath, targetPath); err != nil {
			return reopenCurrent(err)
		}
	} else {
		if err := settingsdb.ResetAtPath(targetPath, targetLocation); err != nil {
			return reopenCurrent(err)
		}
	}

	newStore, err := settingsdb.Open(targetPath)
	if err != nil {
		return reopenCurrent(err)
	}
	if err := newStore.Set(uiprefs.PrefCatalogDBLocation, targetLocation); err != nil {
		_ = newStore.Close()
		return reopenCurrent(err)
	}

	ui.store = newStore
	ui.dbPath = targetPath
	ui.dbLocation = targetLocation
	return nil
}
