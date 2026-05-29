# MANUAL - MSX DB Down

Guia prático para baixar, compilar, executar e operar o sistema.

## Documentos relacionados

- [README](README.md) - visao geral do projeto
- [REFERENCE](REFERENCE.md) - resumo rapido em 5 linguas (menu, funcoes e atalhos)
- [OUTLINE](OUTLINE.md) - contexto de handoff para outra IA/dev
- [CHANGELOG](CHANGELOG.md) - historico de alteracoes por versao

## 1. Visao geral

O **MSX DB Down** e um aplicativo desktop em Go com interface Fyne.

Funcoes atualmente disponiveis:

- Janela principal com area de preview e painel de log/status.
- Menu completo:
  - `File -> Exit`
  - `Setup -> Config UI`
  - `Help -> About`
- Configuracao de UI:
  - tema (`System`, `Light`, `Dark`)
  - fonte (familia)
  - tamanho da fonte
  - densidade de layout
- Internacionalizacao (5 idiomas):
  - `pt`, `en`, `es`, `nl`, `it`
- CLI com Cobra (`--help`, `version`, flags de inicializacao).
- Persistencia em SQLite de configuracoes (`settings.db`).

Versao atual do app: **0.0.3**.

---

## 2. Requisitos

### 2.1 Ferramentas e stack

- Go
- GCC
- CGO/GCO habilitado
- Git
- (Opcional) GoLand
- Runtime/Libs do Fyne
- SQLite (driver embutido: `modernc.org/sqlite`)
- openMSX (integracao futura)

### 2.2 Ambientes usados no projeto

- **Windows 11** com **PowerShell**
- **Fedora 44** com **ZSH**

---

## 3. Como baixar o projeto

### Windows (PowerShell)

```powershell
git clone <URL_DO_REPOSITORIO>
Set-Location "msxdbdown"
go mod tidy
```

### Fedora 44 (ZSH)

```bash
git clone <URL_DO_REPOSITORIO>
cd msxdbdown
go mod tidy
```

> Substitua `<URL_DO_REPOSITORIO>` pela URL real do seu repositório.

---

## 4. Como compilar e executar

### 4.1 Execucao direta (modo desenvolvimento)

#### Windows (PowerShell)

```powershell
Set-Location "C:\dos\msxdbdown"
go run .
```

#### Fedora 44 (ZSH)

```bash
cd /caminho/para/msxdbdown
go run .
```

### 4.2 Compilar binario direto com Go

#### Windows

```powershell
Set-Location "C:\dos\msxdbdown"
go build -o .\dist\windows\msxdbdown.exe .
```

#### Linux

```bash
cd /caminho/para/msxdbdown
go build -o ./dist/linux/msxdbdown .
```

---

## 5. Script de build (`build.ps1`)

O arquivo `build.ps1` encapsula build release/debug, alvos Windows/Linux, injecao de versao/build e execucao opcional do binario.

### 5.1 Parametros disponiveis

- `-Windows`
  Compila para Windows (`GOOS=windows`).

- `-Linux`
  Compila para Linux (`GOOS=linux`).

- `-All`
  Compila para Windows e Linux.

- `-Release`
  Build de release (`-ldflags "-s -w ..."`).

- `-DebugBuild`
  Build de debug (`-gcflags "all=-N -l"`).

- `-Run`
  Executa o binario da plataforma nativa apos compilar.

- `-RunArgs <args...>`
  Passa argumentos para o executavel ao usar `-Run`.

- `-Clean`
  Limpa pasta `dist` antes de compilar.

- `-Version "X.Y.Z"`
  Define versao injetada no binario (ex.: `0.0.3`).

### 5.2 Exemplos de uso do script

#### Release Windows

```powershell
Set-Location "C:\dos\msxdbdown"
.\build.ps1 -Windows -Release -Version "0.0.3"
```

#### Debug Windows + executar comando `version`

```powershell
Set-Location "C:\dos\msxdbdown"
.\build.ps1 -Windows -DebugBuild -Version "0.0.3" -Run -RunArgs "version"
```

#### Build de todos os alvos

```powershell
Set-Location "C:\dos\msxdbdown"
.\build.ps1 -All -Release -Version "0.0.3"
```

#### Limpar artefatos

```powershell
Set-Location "C:\dos\msxdbdown"
.\build.ps1 -Clean
```

### 5.3 Metadados de build injetados

O script injeta no binario:

- `AppVersion`
- `BuildDate` (UTC, formato `ddMMyyyy`)
- `BuildTime` (UTC, formato `HHmm`)
- `BuildNumber` (timestamp UTC em hexadecimal)

---

## 6. CLI do executavel final

Executavel principal (exemplos):

- Windows: `dist\windows\msxdbdown.exe`
- Linux: `dist/linux/msxdbdown`

### 6.1 Comandos

- **Sem subcomando**
  Abre a GUI.

- `version`
  Mostra versao e metadados de build.

### 6.2 Flags globais

- `--lang`, `-l`
  Idioma da UI: `pt | en | es | nl | it`

- `--theme`, `-t`
  Tema inicial da UI: `system | light | dark`

- `--debug`, `-d`
  Ativa mensagens extras no painel de log.

### 6.3 Exemplos

#### Mostrar help em portugues

```powershell
Set-Location "C:\dos\msxdbdown"
go run . --lang pt --help
```

#### Executar GUI em espanhol e tema escuro

```powershell
Set-Location "C:\dos\msxdbdown"
go run . --lang es --theme dark
```

#### Ver versao/build

```powershell
Set-Location "C:\dos\msxdbdown"
go run . version
```

---

## 7. Persistencia de configuracoes

Configuracoes sao salvas em SQLite via pacote `internal/settingsdb`.

### 7.1 Local do banco

- Pasta de configuracao do usuario (`UserConfigDir`) em subpasta `msxdbdown`
- Arquivo: `settings.db`

### 7.2 Chaves salvas

- `ui.language`
- `ui.theme`
- `ui.fontName`
- `ui.fontSize`
- `ui.density`

### 7.3 Regras de inicializacao

- Se idioma salvo existir: app inicia naquele idioma.
- Se nao existir idioma no banco: app inicia em **English**.

---

## 8. Funcionalidades de UI disponiveis hoje

### 8.1 Menu `File`

- `Exit`: encerra a aplicacao.

### 8.2 Menu `Setup`

- `Config UI`: abre dialogo para configurar tema/fonte/tamanho/densidade.
- Ao confirmar, as configuracoes sao persistidas no SQLite.

### 8.3 Menu `Help`

- `About`: abre dialogo com:
  - nome do app
  - versao/build
  - data/hora de build
  - copyrights
  - link clicavel para `https://www.cybernostra.com`

### 8.4 Painel de log/status

- Registra eventos da app e mensagens de depuracao.

---

## 9. Testes e verificacao

### 9.1 Testes focados

```powershell
Set-Location "C:\dos\msxdbdown"
go test ./internal/settingsdb -v
go test ./internal/uiprefs -v
```

### 9.2 Compilacao geral

```powershell
Set-Location "C:\dos\msxdbdown"
go build ./...
```

---

## 10. Estrutura de arquivos relevante

- `main.go` - bootstrap da app + CLI + menus + i18n
- `version.go` - variaveis de build
- `build.ps1` - automacao de build/run
- `internal/about/about.go` - dialogo About
- `internal/configui/configui.go` - dialogo Config UI
- `internal/settingsdb/settingsdb.go` - SQLite de configuracoes
- `internal/uiprefs/uiprefs.go` - defaults e validacao de preferencias
- `internal/uitheme/uitheme.go` - tema customizado
- `README.md` - resumo rapido do projeto
- `OUTLINE.md` - handoff tecnico detalhado para outra IA/dev

---

## 11. Limites atuais e proxima fase

Ainda nao implementado nesta fase:

- downloader das bases externas,
- schema de catalogo (alem de settings),
- matching/normalizacao de nomes,
- enriquecimento de metadados externos,
- launcher integrado com openMSX.

Este manual cobre o estado funcional atual para build, execucao e continuidade do desenvolvimento.

