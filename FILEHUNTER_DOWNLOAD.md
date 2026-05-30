# File-Hunter Download - Guia de Uso

## 📥 Funcionalidades Implementadas

O File-Hunter agora pode fazer download de **dois arquivos** automaticamente:
- `allfiles.txt` - lista completa de arquivos
- `sha1sums.txt` - checksums SHA1 para validação

## 🚀 Como Usar

### Via GUI (Recomendado)

1. **Abra a aplicação**:
   ```powershell
   Set-Location "E:\msxdbdown"
   .\dist\windows\msxdbdown.exe
   ```

2. **Navegue para o menu "Banco de Dados"**:
   - Clique em `Banco de Dados` → `Atualizar File-Hunter`

3. **Você verá na tela principal**:
   - **Endereço**: `https://download.file-hunter.com/allfiles.txt`
   - **SHA URL**: `https://download.file-hunter.com/sha1sums.txt`
   - Botão **Atualizar**

4. **Clique em "Atualizar"**:
   - A status bar ficará **AZUL** (INFO)
   - Mensagem: `[INFO] Starting download to download/ ...`
   - Os dois arquivos serão baixados para a pasta `download/`
   - A status bar mudará para **VERDE** (SUCCESS)
   - Mensagem final: `[INFO] Download completed: download/allfiles.txt, download/sha1sums.txt`

### Recuros Implementados

#### ✅ Dual Download
- Baixa `allfiles.txt` e `sha1sums.txt` em sequência
- Ambos salvos em `./download/`

#### ✅ Retry Automático
- Tenta **2 vezes** se houver falha de rede
- Status muda para **AMARELO** (WARN) em retry
- Mensagem: `[WARN] Retrying download (2/2): ...`

#### ✅ Tratamento de Erros
- Status muda para **VERMELHO** (ERROR) se falhar
- Caixa de diálogo mostra detalhes do erro
- Log completo no painel inferior

#### ✅ URLs Editáveis
- `Setup → Config UI` permite editar os endpoints
- Fallback automático se URL estiver vazia
- Status mostra `[WARN]` quando usando default

#### ✅ Log Estruturado
- Cada download registra:
  - `[HH:MM:SS] [SEVERITY] [DB] Mensagem`
  - Exemplo: `[21:39:14] [INFO] [DB] Download completed: download/allfiles.txt`

## 📊 Estrutura de Diretório

```
msxdbdown/
  ├── main.go                    (UI, menu, download)
  ├── internal/
  │   ├── configui/             (Config UI com URLs)
  │   ├── uiprefs/              (Constantes e defaults)
  │   ├── settingsdb/           (Persistência SQLite)
  │   └── uitheme/              (Contrast visual)
  └── download/                 (Pasta criada ao baixar)
      ├── allfiles.txt
      └── sha1sums.txt
```

## 🔧 Testando em Desenvolvimento

```powershell
# Build e rodar
Set-Location "E:\msxdbdown"
go run . --lang pt

# Build release
.\build.ps1 -Windows -Release -Version "0.1.7" -Run
```

## 📝 Próximos Passos (Sugestão)

1. **Validação SHA1**: Comparar sha1sums.txt com arquivos baixados
2. **Extração**: Deszipar/descompactar allfiles.txt se necessário
3. **Barra de Progresso**: Mostrar % do download em tempo real
4. **Cancelamento**: Permitir parar o download em progresso

---

**Status**: ✅ Pronto para uso
**Última alteração**: 29/05/2026
**Severidades**: INFO (azul), WARN (amarelo), ERROR (vermelho)

