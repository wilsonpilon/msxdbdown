# 📥 Download do File-Hunter - Resumo Técnico

## Fluxo Implementado

```
┌─────────────────────────────────────────────┐
│  Menu: Banco de Dados → Atualizar File-Hunter │
└───────────────┬─────────────────────────────┘
                │
                ↓
    ┌─────────────────────────────┐
    │ showFileHunterUpdateView()   │
    │ • Lê URLs do SQLite          │
    │ • Se vazia → WARN + default  │
    │ • Mostra 2 campos (URL)      │
    │ • Botão "Atualizar"          │
    └───────────┬─────────────────┘
                │
                ↓
    ┌─────────────────────────────────────┐
    │ Clique em "Atualizar"               │
    │ resolveURLSetting() x2              │
    │ • Checa se vazio                    │
    │ • Log WARN se fallback              │
    │ • Retorna URL resolvida             │
    └───────────┬─────────────────────────┘
                │
                ↓
    ┌──────────────────────────────────────────┐
    │ downloadSingleFile(url) - 1ª vez        │
    │ • Parse URL + validação                 │
    │ • Cria download/ se não existe          │
    │ • HTTP GET com timeout 180s             │
    │ • Salva em .part (atomic)               │
    └───────────┬──────────────────────────────┘
                │
        ┌───────┴────────┐
        │                │
    OK? ↓                ↓ ERRO
  succ  │              retry (attempt 2/2)
        │    ┌──────────────────────────────┐
        │    │ Log WARN: Retrying... (2/2)  │
        │    │ HTTP GET novamente           │
        │    └──────────┬───────────────────┘
        │               │
        │               ├── OK? → rename .part → final
        │               └── ERRO? → return err
        │
        ├→ rename .part → finalPath
        ├→ Log INFO: "Download completed: path"
        └→ Retorna path

                ↓
    ┌──────────────────────────────────┐
    │ 2º arquivo (sha1sums.txt)         │
    │ Mesmo fluxo acima                 │
    └──────────────────────────────────┘
                │
                ↓
    ┌──────────────────────────────────┐
    │ Status Bar - Cores                │
    │ INFO  → Azul   (#3A78C2)          │
    │ WARN  → Âmbar  (#CF8A16)          │
    │ ERROR → Vermelho (#C62D42)        │
    └──────────────────────────────────┘
```

## Arquivos Envolvidos

| Arquivo | Função |
|---------|--------|
| `main.go` | Menu, download, retry, severidade |
| `internal/configui/configui.go` | Edição de URLs no Setup |
| `internal/uiprefs/uiprefs.go` | Constantes e defaults de URL |
| `internal/settingsdb/settingsdb.go` | Persistência das URLs em SQLite |
| `internal/uitheme/uitheme.go` | Contraste visual (foreground/background) |

## Métodos Principais

### `showFileHunterUpdateView()` (main.go)
- Renderiza tela com dois campos de URL
- Botão "Atualizar" dispara downloads sequenciais
- Mostra severidade do status em tempo real

### `downloadSingleFile(sourceURL string)` (main.go)
- **Parâmetros**: URL do arquivo a baixar
- **Retorno**: caminho final ou erro
- **Lógica**:
  - Tenta 2 vezes (retry automático)
  - Usa `.part` temporário (atomic write)
  - Valida HTTP 200-299
  - Registra WARN em retries
  - Registra ERROR em falhas finais

### `resolveURLSetting(key, fallback, label)` (main.go)
- **Parâmetros**: chave SQLite, URL default, rótulo
- **Retorno**: URL resolvida
- **Lógica**:
  - Se vazio → registra WARN + usa fallback
  - Se preenchido → retorna como está

### `setStatus(severity, message)` (main.go)
- Atualiza texto + cor da status bar
- Cores: INFO/WARN/ERROR

## Constantes Definidas

```go
// URLs padrão (em internal/uiprefs/uiprefs.go)
DefaultFileHunterURL    = "https://download.file-hunter.com/allfiles.txt"
DefaultFileHunterSHAURL = "https://download.file-hunter.com/sha1sums.txt"

// Chaves SQLite
PrefFileHunterURL    = "db.filehunter.url"
PrefFileHunterSHAURL = "db.filehunter.sha.url"

// Severidades
statusInfo  = "INFO"
statusWarn  = "WARN"
statusError = "ERROR"
```

## Exemplo de Log

```
[21:39:14] [INFO] [DB] File-Hunter update view opened
[21:39:14] [WARN] [DB] File-Hunter URL is empty; using default: https://download.file-hunter.com/allfiles.txt
[21:39:14] [WARN] [DB] File-Hunter SHA URL is empty; using default: https://download.file-hunter.com/sha1sums.txt
[21:39:14] [INFO] [DB] Starting download to download/ ...
[21:39:25] [INFO] [DB] Download completed: ./download/allfiles.txt, ./download/sha1sums.txt
```

## Status Esperado

- ✅ Ambos arquivos baixados para `./download/`
- ✅ Log com timestamps, severidade, scope
- ✅ Retry automático em falhas de rede
- ✅ URLs editáveis no Setup
- ✅ Fallback para default se vazio
- ✅ Status bar: cor + texto + bold
- ✅ Sem dependências externas (stdlib + Fyne + Cobra + SQLite)

---
**Build**: ✓ OK (09/05/2026 21:39)
**Status**: ✓ Pronto para uso

