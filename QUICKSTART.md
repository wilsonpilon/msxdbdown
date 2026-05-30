# 🚀 Quick Start - File-Hunter Download

## ⚡ 30 segundos de setup

```powershell
cd E:\msxdbdown
go run . --lang pt
```

## 📥 Download em 3 cliques

1. **Banco de Dados** → **Atualizar File-Hunter**
2. Veja os 2 URLs (editáveis em Setup)
3. Clique **Atualizar** → arquivos em `./download/`

## 📊 O que você verá

### Status Bar (rodapé)
```
[INFO] Download completed: ./download/allfiles.txt, ./download/sha1sums.txt
```

Cores:
- 🔵 Azul = OK / INFO
- 🟨 Âmbar = Retry / WARN
- 🔴 Vermelho = Erro / ERROR

### Log Panel (canto inferior)
```
[21:39:14] [INFO] [DB] Starting download to download/ ...
[21:39:25] [INFO] [DB] Download completed: ./download/allfiles.txt, ./download/sha1sums.txt
```

## 🔧 Customizar URLs

Menu → **Setup → Config UI** → edite os campos JSON

```
MSX ROM DB URL        = https://romdb.vampier.net/Archive/sql-msxromdb.zip
File-Hunter URL       = https://download.file-hunter.com/allfiles.txt
File-Hunter SHA URL   = https://download.file-hunter.com/sha1sums.txt
```

Clique **Apply** → salvo no SQLite

## 📁 Arquivos gerados

```
./download/
├── allfiles.txt      (~1-5 MB)
└── sha1sums.txt      (~100 KB)
```

## ❌ Se algo falhar

- **"Retrying download (2/2)"** → 2ª tentativa automática
- **RED status** → Ver diálogo de erro (detalhe tech)
- **Vazio no log** → Check internet connection

---

✅ **Pronto!** File-Hunter download agora funciona.

