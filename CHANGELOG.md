# Changelog

Todas as alteracoes relevantes deste projeto serao registradas neste arquivo.

O formato e inspirado em Keep a Changelog, com versionamento semantico.

## [0.1.7] - 2026-05-30

### Added

- Menu `Banco de Dados` expandido com fluxos ativos para:
  - `Atualizar MSX RomDB`
  - `Atualizar File-Hunter`
  - `Limpar Downloads`
- Importacao automatica do dump SQL do MSX RomDB para o SQLite atual da aplicacao.
- Aba `SQLite` em `Setup -> Config UI` com:
  - visualizacao do caminho atual do banco,
  - selecao entre banco local e pasta de configuracao do usuario,
  - criacao de banco inicial com defaults,
  - opcao para mover banco atual ou criar novo banco zerado.
- Logs detalhados por tabela durante a importacao SQL:
  - backup criado,
  - tabela recriada,
  - backup removido,
  - resumo final com contadores agregados.

### Changed

- Banco de configuracoes padrao/documentado passou a ser `msxdbdown.db`.
- Localizacao do banco unificada para:
  - `data/msxdbdown.db`
  - `%APPDATA%/msxdbdown/msxdbdown.db` no Windows
  - `~/.config/msxdbdown/msxdbdown.db` no Linux
- Importacao do MSX RomDB evoluiu de simples insercao para refresh atomico por tabela.
- Documentacao principal atualizada para refletir o estado funcional atual do projeto.

### Fixed

- Reimportacao do MSX RomDB nao duplica mais registros no refresh.
- Janela de inconsistencia durante refresh de tabelas grandes reduzida com estrategia atomica por tabela.

### Tests

- Testes de `internal/settingsdb` expandidos para cobrir:
  - defaults em banco novo,
  - importacao SQL,
  - refresh sem duplicacao,
  - refresh atomico por tabela,
  - logs detalhados e resumo final.

## [0.0.3] - 2026-05-29

### Added

- Bootstrap completo da GUI com **Fyne**:
  - janela principal com painel lateral de preferencias,
  - area principal (placeholder do catalogo),
  - painel de log/status,
  - barra de status inferior.
- Menus principais com estrutura funcional:
  - `File -> Exit`
  - `Setup -> Config UI`
  - `Help -> About`
- Suporte a 5 idiomas na UI:
  - Portugues (`pt`), English (`en`), Espanol (`es`), Nederlands (`nl`), Italiano (`it`).
- Selecao de tema:
  - `System`, `Light`, `Dark`.
- Dialogo **Config UI**:
  - familia de fonte,
  - tamanho de fonte,
  - densidade de layout.
- Dialogo **About** com:
  - nome do app,
  - versao/build,
  - data e hora da build,
  - copyrights,
  - link clicavel para `https://www.cybernostra.com`.
- CLI com **Cobra**:
  - comando padrao abre GUI,
  - subcomando `version`,
  - flags `--lang`, `--theme`, `--debug`.
- Localizacao dos textos da CLI (Short, Long, flags e resumo do comando version) de acordo com `--lang`.
- Persistencia de configuracoes em **SQLite** (driver `modernc.org/sqlite`) via novo pacote `internal/settingsdb`.
- Script de build `build.ps1` com:
  - alvos `-Windows`, `-Linux`, `-All`,
  - perfis `-Release` e `-DebugBuild`,
  - execucao opcional `-Run`/`-RunArgs`,
  - limpeza `-Clean`,
  - injecao de metadados de build (`AppVersion`, `BuildDate`, `BuildTime`, `BuildNumber`).
- Documentacao adicional:
  - `OUTLINE.md` (handoff tecnico para continuidade em outro ambiente),
  - `MANUAL.md` (download/build/run/opcoes de CLI),
  - screenshot no `README.md` (`images/msxdbdown-00.png`).

### Changed

- Menus da GUI traduzidos para as 5 linguas.
- Fallback de idioma alterado para **English** quando nao houver valor salvo.
- `Config UI` deixou de gravar em `fyne.Preferences` e passou a gravar em SQLite.
- Inicializacao da app ajustada para carregar configuracoes do SQLite (idioma/tema/fonte/tamanho/densidade).
- About refinado visualmente (iteracoes de layout, agrupamento de linhas e centralizacao de texto de conteudo).
- README reestruturado com versao atual, stack e ambientes (Windows 11 + PowerShell / Fedora 44 + ZSH).

### Fixed

- Crash de inicializacao da GUI por callback de `Select` disparado antes da UI estar totalmente montada.
- Problemas de parsing de `-ldflags` no build quando havia metadado com espaco.
  - Resolvido separando `BuildDate` e `BuildTime`.
- Ajustes em assinatura/chamada do About para exibir corretamente metadados de versao/build.

### Tests

- Testes de `internal/uiprefs` atualizados para refletir novo fallback de idioma (ingles).
- Testes novos em `internal/settingsdb` cobrindo `Set/Get` e chave inexistente.
- Verificacoes recorrentes de `go build ./...` e suites de testes internas durante as alteracoes.

---

## Como manter este changelog nas proximas versoes

Ao iniciar uma nova versao:

1. Crie uma nova secao no topo no formato:
   - `## [X.Y.Z] - YYYY-MM-DD`
2. Resuma por categoria:
   - `Added`, `Changed`, `Fixed`, `Tests`.
3. Foque no impacto funcional (o que foi entregue), nao apenas em arquivos alterados.
4. Ao fazer release, mantenha um sumario curto da direcao da versao.

Exemplo de template:

```markdown
## [0.0.4] - 2026-06-XX

### Added
- ...

### Changed
- ...

### Fixed
- ...

### Tests
- ...
```

