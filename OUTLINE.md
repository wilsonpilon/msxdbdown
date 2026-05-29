# MSX DB Down - Project Outline / Handoff

Este arquivo resume o que foi pedido e implementado durante o chat, para outra IA/dev conseguir continuar em outro computador/sistema.

## 1) Objetivo do Projeto

Criar um aplicativo desktop em Go para Windows/Linux, com GUI em Fyne, para:

- baixar duas bases de dados de sites especificos,
- consolidar em SQLite,
- analisar nomes similares para padronizacao entre bases,
- enriquecer com metadados externos (imagem, video, musica, lancamento, fabricante etc.),
- montar um catalogo visual que tambem funcione como frontend para emulador MSX.

## 2) Escopo ja implementado

A etapa atual foca no bootstrap da aplicacao (UI + CLI + build + persistencia de configuracoes).

### 2.1 UI principal

Implementado em `main.go`:

- Janela principal com layout moderno:
  - painel lateral de preferencias,
  - area principal (placeholder do catalogo visual),
  - painel de log/status para progresso/depuracao,
  - status bar inferior.
- Menu principal:
  - `File -> Exit` (funcional)
  - `Setup -> Config UI` (funcional)
  - `Help -> About` (funcional)
- Troca de idioma em runtime (5 idiomas):
  - Portugues (`pt`)
  - English (`en`)
  - Espanol (`es`)
  - Nederlands (`nl`)
  - Italiano (`it`)
- Troca de tema:
  - System
  - Light
  - Dark

### 2.2 Dialogo Config UI

Implementado em `internal/configui/configui.go`:

- Configuracao de familia de fonte,
- tamanho de fonte,
- densidade de layout,
- persistencia ao clicar Apply,
- callback para reaplicar tema imediatamente.

### 2.3 Dialogo About

Implementado em `internal/about/about.go`:

- titulo "MSX DB Down" com destaque,
- versao + build na mesma linha,
- data + hora de build na mesma linha,
- copyrights na mesma linha,
- link clicavel para `https://www.cybernostra.com`,
- ano dinamico `1972 - <ano atual>`.

### 2.4 Menu traduzido nas 5 linguas

As chaves de menu foram traduzidas no i18n em `main.go`:

- `menuFile`, `menuExit`, `menuSetup`, `menuConfigUI`, `menuHelp`, `menuAbout`

### 2.5 CLI com Cobra

Implementado em `main.go`:

- comando raiz abre GUI,
- subcomando `version`,
- flags:
  - `--lang`, `-l`
  - `--theme`, `-t`
  - `--debug`, `-d`
- help do Cobra localizado conforme `--lang`:
  - `Short`, `Long`, descricao de flags e `version short` traduzidos.

## 3) Build / versionamento da build

### 3.1 Variaveis de build

Em `version.go`:

- `AppVersion`
- `BuildDate`
- `BuildTime`
- `BuildNumber`

Valores injetados via `-ldflags` no build.

### 3.2 Script de build

Em `build.ps1`:

- targets: `-Windows`, `-Linux`, `-All`
- perfil: `-Release` ou `-DebugBuild`
- execucao opcional: `-Run -RunArgs ...`
- limpeza: `-Clean`
- versao: `-Version "0.0.3"`

Build metadata:

- `BuildDate`: formato militar `ddMMyyyy` (UTC)
- `BuildTime`: formato militar `HHmm` (UTC)
- `BuildNumber`: timestamp UTC convertido para hexadecimal

## 4) Persistencia em SQLite (pedido recente)

### 4.1 O que foi feito

Foi adicionada uma camada SQLite para guardar configuracoes de UI e idioma:

- novo pacote: `internal/settingsdb/settingsdb.go`
- driver: `modernc.org/sqlite` (pure Go)
- banco: `settings.db` em `%UserConfigDir%/msxdbdown/` (ou equivalente do OS)
- tabela:

```sql
CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
```

### 4.2 Chaves salvas no SQLite

- `ui.language` (ultima linguagem usada)
- `ui.theme`
- `ui.fontName`
- `ui.fontSize`
- `ui.density`

### 4.3 Comportamento de idioma default

- Se `ui.language` nao existir/for invalido no banco: default agora e `English` (`en`).
- Isso foi ajustado em `internal/uiprefs/uiprefs.go` (`ReadLanguage`).

### 4.4 Integracao com UI

`main.go` foi alterado para:

- abrir `settingsdb.OpenDefault()` no startup,
- usar callbacks `getSetting/setSetting` para leitura/escrita,
- passar esses callbacks para `configui.Show(...)`,
- carregar e aplicar tema/fonte/densidade do SQLite em `applyAll()`,
- persistir troca de idioma e tema no SQLite.

## 5) Estrutura atual de pacotes internos

- `internal/about/` - dialogo About
- `internal/configui/` - dialogo Config UI
- `internal/settingsdb/` - persistencia SQLite de settings
- `internal/uiprefs/` - normalizacao/defaults de idioma/tema/fonte/densidade
- `internal/uitheme/` - tema custom Fyne (base + overrides)

## 6) Dependencias principais

`go.mod` usa:

- `fyne.io/fyne/v2 v2.7.4`
- `github.com/spf13/cobra v1.10.2`
- `modernc.org/sqlite v1.51.0`

## 7) Testes existentes

- `internal/uiprefs/uiprefs_test.go`
- `internal/settingsdb/settingsdb_test.go`

Cobrem:

- defaults/validacao de idioma/tema/fonte/densidade,
- set/get no SQLite e retorno vazio para chave inexistente.

## 8) Comandos uteis

### 8.1 Rodar app

```powershell
Set-Location "C:\dos\msxdbdown"
go run .
```

### 8.2 Help da CLI em idioma especifico

```powershell
Set-Location "C:\dos\msxdbdown"
go run . --lang pt --help
```

### 8.3 Testes focados

```powershell
Set-Location "C:\dos\msxdbdown"
go test ./internal/settingsdb -v
go test ./internal/uiprefs -v
```

### 8.4 Build via script

```powershell
Set-Location "C:\dos\msxdbdown"
.\build.ps1 -Windows -Release -Version "0.0.3"
```

```powershell
Set-Location "C:\dos\msxdbdown"
.\build.ps1 -Windows -DebugBuild -Version "0.0.3" -Run -RunArgs "version"
```

## 9) Historico resumido de decisoes relevantes

1. Inicialmente, configuracoes ficavam em `fyne.Preferences`.
2. Foi migrado para SQLite para persistencia unificada de Setup/Config UI e idioma.
3. Idioma default foi alterado para `English` quando nao houver valor no banco.
4. O `About` passou por iteracoes visuais:
   - layout mais largo/baixo,
   - informacoes consolidadas por linha,
   - link clicavel,
   - ajustes de alinhamento/estilo.
5. Menus e CLI help foram internacionalizados para 5 idiomas.

## 10) Pendencias / proximos passos (nao implementados ainda)

Escopo macro original ainda pendente:

- downloader das 2 fontes externas,
- schema SQLite para dados de catalogo (alem de settings),
- algoritmo de matching/normalizacao de nomes,
- scraping/API para metadados (imagem/video/musica etc.),
- tela de catalogo visual real (lista/cards/filtros/detalhes),
- integracao launcher com emulador MSX.

## 11) Observacoes para outra IA/dev

- Priorize manter compatibilidade Windows + Linux.
- Evite quebrar `build.ps1` (pipeline local ja usado).
- Preservar internacionalizacao existente em `main.go`.
- Ao tocar persistencia de settings, manter fallback robusto para defaults.
- Se evoluir banco principal de catalogo, mantenha `settings.db` separado (ou migre com cuidado).

