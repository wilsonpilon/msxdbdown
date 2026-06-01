# MSX DB Down - Project Outline / Handoff

Este arquivo resume o que foi pedido e implementado durante o chat, para outra IA/dev conseguir continuar em outro computador/sistema.

## 1) Objetivo do Projeto

Criar um aplicativo desktop em Go para Windows/Linux, com GUI em Fyne, para:

- baixar bases de dados de sites especificos,
- consolidar em SQLite,
- analisar nomes similares para padronizacao entre bases,
- enriquecer com metadados externos (imagem, video, musica, lancamento, fabricante etc.),
- montar um catalogo visual que tambem funcione como frontend para emulador MSX.

## 2) Escopo ja implementado

A etapa atual foca no bootstrap da aplicacao, configuracao persistente, downloads de bases e importacao inicial para SQLite.

### 2.0 Database Download Menu

Implementado em `main.go`:

- Menu `Banco de Dados` com submenus:
  - `Atualizar MSX RomDB` - baixa zip do dump SQL, descompacta e importa no SQLite atual
  - `Atualizar File-Hunter` - baixa 2 arquivos (`allfiles.txt` + `sha1sums.txt`)
  - `Browse File-Hunter` - navega no catalogo File-Hunter importado no SQLite
  - `Limpar Downloads` - remove arquivos em `download/`
- Tela de download mostrando URL(s) + botao `Atualizar`
- Download com retry automatico (2 tentativas)
- Fallback para URL default se campo vazio
- Status com cores por severidade:
  - INFO (azul): eventos normais
  - WARN (ambar): retries, fallbacks de URL
  - ERROR (vermelho): falhas finais
- Importacao do MSX RomDB com:
  - refresh atomico por tabela,
  - logs detalhados por tabela,
  - resumo final com contadores.

### 2.0.1 Catalogo File-Hunter (novo)

Implementado em `internal/settingsdb/settingsdb_filehunter.go` e integrado em `main.go`:

- Importacao de `download/allfiles.txt` para schema normalizado no SQLite.
- Importacao de `download/sha1sums.txt` para atualizar SHA1 dos arquivos importados.
- Tabelas novas:
  - `fh_category`
  - `fh_file_type`
  - `fh_file`
  - `fh_file_category`
- Regras de parsing:
  - `full_path` preserva caminho completo (ex.: `GAMES\\MSX1\\ROM\\Gradius 2 ...zip`)
  - `name` guarda nome sem extensao (chave textual de busca)
  - extensao normalizada (ex.: `zip`) em tabela dedicada de tipo
  - categorias hierarquicas registradas por `position` (nivel no caminho)
- Busca/navegacao prontas no Store:
  - `ListFHCategories(pathFilter)`
  - `ListFHFileTypes(pathFilter)`
  - `SearchFHFiles(pathFilter, nameQuery, extension, limit)`

### 2.0.2 Browse File-Hunter (UI)

Implementado em `main.go` (`showFileHunterBrowseView`):

- Painel de categorias com contagem e navegacao por niveis.
- Breadcrumb clicavel para subir no caminho.
- Filtro por extensao + busca por nome.
- Tabela com `Nome`, `Ext`, `SHA1` e `Caminho completo`.
- Clique na linha copia SHA1 para a area de transferencia.

### 2.1 UI principal

Implementado em `main.go`:

- Janela principal com layout moderno:
  - painel lateral de preferencias,
  - area principal (placeholder do catalogo visual),
  - painel de log/status para progresso/depuracao,
  - status bar inferior.
- Menu principal:
  - `File -> Exit`
  - `Setup -> Config UI`
  - `Help -> About`
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

- Dialogo paginado em abas:
  - UI
  - URLs
  - SQLite
- Configuracao de familia de fonte,
- tamanho de fonte,
- densidade de layout,
- URLs de MSX RomDB e File-Hunter,
- visualizacao do caminho atual do banco,
- alternancia entre `data/msxdbdown.db` e pasta de configuracao do usuario,
- opcao para mover banco atual ou criar novo banco zerado,
- callback para reaplicar tema imediatamente.

### 2.3 Dialogo About

Implementado em `internal/about/about.go`:

- titulo "MSX DB Down" com destaque,
- versao + build na mesma linha,
- data + hora de build na mesma linha,
- copyrights na mesma linha,
- link clicavel para `https://www.cybernostra.com`,
- ano dinamico `1972 - <ano atual>`,
- exibe o icone do aplicativo no topo do dialogo.

### 2.3.1 Icone do aplicativo

Implementado em `internal/appicon/appicon.go`:

- Recurso de icone monocromatico (MSX + database) gerado em runtime como `fyne.Resource`.
- Aplicado em:
  - icone global da app (`application.SetIcon(...)`),
  - icone da janela principal (`window.SetIcon(...)`),
  - imagem no dialogo About.

### 2.4 Menu traduzido nas 5 linguas

As chaves de menu e telas foram traduzidas no i18n em `main.go`.

### 2.5 CLI com Cobra

Implementado em `main.go`:

- comando raiz abre GUI,
- subcomando `version`,
- flags:
  - `--lang`, `-l`
  - `--theme`, `-t`
  - `--debug`, `-d`
- help do Cobra localizado conforme `--lang`.

## 3) Build / versionamento da build

### 3.1 Variaveis de build

Em `version.go`:

- `AppVersion`
- `BuildDate`
- `BuildTime`
- `BuildNumber`

Valores injetados via `-ldflags` no build.

Versao atual do projeto: **0.1.7**.

### 3.2 Script de build

Em `build.ps1`:

- targets: `-Windows`, `-Linux`, `-All`
- perfil: `-Release` ou `-DebugBuild`
- execucao opcional: `-Run -RunArgs ...`
- limpeza: `-Clean`
- versao: `-Version "0.1.7"`
- icone no EXE Windows embutido automaticamente via `rsrc`
- opcao `-NoIcon` para pular geracao/embedding de icone (build mais rapido)

Default atual no script:

- `-Version` padrao = `0.1.7`

Build metadata:

- `BuildDate`: formato militar `ddMMyyyy` (UTC)
- `BuildTime`: formato militar `HHmm` (UTC)
- `BuildNumber`: timestamp UTC convertido para hexadecimal

## 4) Persistencia em SQLite

### 4.1 O que foi feito

Foi adicionada uma camada SQLite para guardar configuracoes de UI, URLs e localizacao do banco:

- pacote: `internal/settingsdb/settingsdb.go`
- driver: `modernc.org/sqlite` (pure Go)
- banco principal atual: `msxdbdown.db`
- localizacoes suportadas:
  - `data/msxdbdown.db`
  - `%APPDATA%/msxdbdown/msxdbdown.db` (Windows)
  - `~/.config/msxdbdown/msxdbdown.db` (Linux)
- tabela base de settings:

```sql
CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
```

### 4.2 Chaves salvas no SQLite

- `ui.language`
- `ui.theme`
- `ui.fontName`
- `ui.fontSize`
- `ui.density`
- `db.msxromdb.url`
- `db.filehunter.url`
- `db.filehunter.sha.url`
- `db.catalog.location`

### 4.3 Comportamento default

- Se `ui.language` nao existir/for invalido: default = `English` (`en`).
- Banco novo recebe preferencias padrao automaticamente.

### 4.4 Integracao com UI

`main.go` foi alterado para:

- detectar o banco ativo no startup,
- abrir o SQLite correto no startup,
- usar callbacks `getSetting/setSetting` para leitura/escrita,
- passar callbacks extras para `configui.Show(...)`,
- carregar e aplicar tema/fonte/densidade do SQLite em `applyAll()`,
- persistir troca de idioma e tema no SQLite,
- trocar o banco ativo em runtime quando o usuario mudar a localizacao.

## 5) Importacao do MSX RomDB

Implementado em `internal/settingsdb/settingsdb.go` e integrado em `main.go`.

Fluxo atual:

1. Download de `sql-msxromdb.zip`
2. Descompactacao do `.sql` em `download/`
3. Importacao no SQLite atual
4. Refresh atomico por tabela:
   - backup da tabela existente,
   - recriacao da tabela,
   - insercao dos novos dados,
   - remocao do backup apenas ao final com sucesso.
5. Logs emitidos:
   - backup criado,
   - tabela recriada,
   - backup removido,
   - resumo final com contadores.

## 6) Estrutura atual de pacotes internos

- `internal/about/` - dialogo About
- `internal/appicon/` - geracao do icone monocromatico da app
- `internal/configui/` - dialogo Config UI com abas UI / URLs / SQLite
- `internal/settingsdb/` - persistencia SQLite + importacao SQL + catalogo File-Hunter
- `internal/uiprefs/` - normalizacao/defaults de idioma/tema/fonte/densidade + URLs
- `internal/uitheme/` - tema custom Fyne

## 7) Dependencias principais

`go.mod` usa:

- `fyne.io/fyne/v2 v2.7.4`
- `github.com/spf13/cobra v1.10.2`
- `modernc.org/sqlite v1.51.0`

## 8) Testes existentes

- `internal/uiprefs/uiprefs_test.go`
- `internal/settingsdb/settingsdb_test.go`

Cobrem:

- defaults/validacao de idioma/tema/fonte/densidade,
- URL fallback quando vazio,
- set/get no SQLite e retorno vazio para chave inexistente,
- seed de defaults em banco novo,
- importacao SQL,
- refresh sem duplicacao,
- refresh atomico por tabela,
- logs detalhados e resumo final.

Observacao de estado atual:

- `go test ./internal/settingsdb` tem uma falha conhecida em `TestGetRomVersionsByGameID`
  por divergencia de schema de teste (`no such column: FileSize`) e nao por falha do fluxo File-Hunter.

## 9) Comandos uteis

### 9.1 Rodar app

```powershell
Set-Location "C:\dos\msxdbdown"
go run .
```

### 9.2 Help da CLI em idioma especifico

```powershell
Set-Location "C:\dos\msxdbdown"
go run . --lang pt --help
```

### 9.3 Testes focados

```powershell
Set-Location "C:\dos\msxdbdown"
go test ./internal/settingsdb -v
go test ./internal/uiprefs -v
go test . ./internal/appicon/
```

### 9.4 Build via script

```powershell
Set-Location "C:\dos\msxdbdown"
.\build.ps1 -Windows -Release -Version "0.1.7"
```

```powershell
Set-Location "C:\dos\msxdbdown"
.\build.ps1 -Windows -DebugBuild -Version "0.1.7" -Run -RunArgs "version"
```

```powershell
Set-Location "C:\dos\msxdbdown"
.\build.ps1 -Windows -DebugBuild -NoIcon -Run
```

## 10) Historico resumido de decisoes relevantes

1. Configuracoes sairam de `fyne.Preferences` para SQLite.
2. Idioma default foi alterado para `English` quando nao houver valor no banco.
3. O `Config UI` foi paginado em abas.
4. O banco de configuracoes passou a poder ficar em pasta local ou pasta de configuracao do usuario.
5. O MSX RomDB agora e importado para o SQLite da app.
6. O refresh da importacao passou de `DELETE FROM` para estrategia atomica por tabela.
7. A importacao ganhou logs por tabela e resumo agregado.
8. Foi adicionado catalogo File-Hunter normalizado no SQLite (categorias, tipo de arquivo, SHA1).
9. O app passou a ter icone proprio aplicado na janela/About e no EXE Windows.
10. O build ganhou modo `-NoIcon` para iteracao rapida.

## 11) Pendencias / proximos passos

Escopo macro original ainda pendente:

- schema consolidado do catalogo final alem das tabelas importadas,
- algoritmo de matching/normalizacao de nomes,
- scraping/API para metadados,
- tela de catalogo visual real,
- integracao launcher com emulador MSX,
- barra de progresso em tempo real.

## 12) Observacoes para outra IA/dev

- Priorize manter compatibilidade Windows + Linux.
- Evite quebrar `build.ps1`.
- Preservar internacionalizacao existente em `main.go`.
- Ao tocar persistencia de settings, manter fallback robusto para defaults.
- Ao mexer na importacao SQL, preservar a estrategia atomica por tabela.
