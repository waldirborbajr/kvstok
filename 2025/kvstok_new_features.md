# 🚀 KVSTOK - ROADMAP DE NOVAS FEATURES

## Feature #1: Search & Filter

### Descrição
Permitir buscar keys por padrão, filtrar por prefixo e exportar em múltiplos formatos.

### Especificação Técnica

```go
// cmd/commands/search.go - NOVO ARQUIVO

var SearchCmd = &cobra.Command{
    Use:     "{s}earch [PATTERN]",
    Short:   "Search for keys matching pattern (regex or glob).",
    Aliases: []string{"s"},
    Args:    cobra.MinimumNArgs(1),
    Flags: map[string]string{
        "--regex":  "Use regex pattern (default: glob)",
        "--prefix": "Match only at start",
        "--json":   "Output as JSON",
    },
    Run: func(cmd *cobra.Command, args []string) {
        pattern := args[0]
        regex, _ := cmd.Flags().GetBool("regex")
        prefix, _ := cmd.Flags().GetBool("prefix")
        jsonOut, _ := cmd.Flags().GetBool("json")
        
        database.DB.View(func(tx *nutsdb.Tx) error {
            keys, values, _ := tx.GetAll(database.Bucket)
            results := make(map[string]string)
            
            for i := 0; i < len(keys); i++ {
                k := string(keys[i])
                if matchesPattern(k, pattern, regex, prefix) {
                    results[k] = string(values[i])
                }
            }
            
            if jsonOut {
                data, _ := json.MarshalIndent(results, "", "  ")
                fmt.Printf("%s\n", data)
            } else {
                for k, v := range results {
                    fmt.Printf("%s\t%s\n", k, v)
                }
            }
            return nil
        })
    },
}

func matchesPattern(key string, pattern string, useRegex bool, prefixOnly bool) bool {
    if useRegex {
        r, _ := regexp.Compile(pattern)
        return r.MatchString(key)
    } else if prefixOnly {
        return strings.HasPrefix(key, pattern)
    } else {
        // Glob pattern
        matched, _ := filepath.Match(pattern, key)
        return matched
    }
}
```

### Exemplos de Uso
```bash
# Buscar chaves com "docker"
kvstok s "*docker*"

# Buscar chaves que começam com "db_"
kvstok s "db_*" --prefix

# Buscar com regex
kvstok s "^prod_.*_token$" --regex

# Saída JSON
kvstok s "aws*" --json > aws-keys.json
```

### Esforço: 2h
### Benefício: Muito Alto - Essencial para >50 keys

---

## Feature #2: Tags/Categories

### Descrição
Associar tags a cada key-value para melhor organização.

### Banco de Dados Estendido
```go
// internal/database/model.go - NOVO

type SecretEntry struct {
    Value  string
    TTL    uint32
    Tags   []string      // ✅ NOVO
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Especificação Técnica

```bash
# Comandos
kvstok add mykey myvalue --tags env,prod,critical
kvstok add anothekey anothervalue -t docker -t staging

# Listar por tag
kvstok list --tag prod         # AND lógico
kvstok list --tag prod --tag db # Múltiplas tags

# Remover/Atualizar tags
kvstok tag add mykey newenv
kvstok tag remove mykey prod
kvstok tag list mykey          # Mostra tags de uma key
```

### Implementação
```go
// cmd/commands/tag.go - NOVO ARQUIVO

var TagCmd = &cobra.Command{
    Use:     "tag [add|remove|list] [KEY] [TAG...]",
    Short:   "Manage tags for a key.",
    Run: func(cmd *cobra.Command, args []string) {
        subcommand := args[0]
        key := args[1]
        
        switch subcommand {
        case "add":
            tags := args[2:]
            database.DB.Update(func(tx *nutsdb.Tx) error {
                // Fetch existing entry
                // Add new tags to existing list
                // Update in DB
                return nil
            })
        case "remove":
            // Similar logic
        case "list":
            // Display tags for key
        }
    },
}

// Modificar list.go para filtrar por tags
func filterByTags(keys [][]byte, values [][]byte, tags []string) [][]byte {
    // Filter logic: buscar entries com TODAS as tags solicitadas
    // Return filtered keys
}
```

### Esforço: 3h
### Benefício: Alto - Organização essencial

---

## Feature #3: Encrypt Exports

### Descrição
Encriptar `kvstok.json` com AES-256-GCM antes de salvar.

### Especificação Técnica

```go
// internal/crypto/export.go - NOVO

package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "io"
)

type ExportEncryptor struct {
    masterKey []byte // Derivada de passphrase via Argon2
}

func (e *ExportEncryptor) EncryptExport(data map[string]string) ([]byte, error) {
    block, _ := aes.NewCipher(e.masterKey)
    gcm, _ := cipher.NewGCM(block)
    
    nonce := make([]byte, gcm.NonceSize())
    io.ReadFull(rand.Reader, nonce)
    
    plaintext, _ := json.Marshal(data)
    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    
    return ciphertext, nil
}

func (e *ExportEncryptor) DecryptExport(data []byte) (map[string]string, error) {
    block, _ := aes.NewCipher(e.masterKey)
    gcm, _ := cipher.NewGCM(block)
    
    nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
    plaintext, _ := gcm.Open(nil, nonce, ciphertext, nil)
    
    var result map[string]string
    json.Unmarshal(plaintext, &result)
    
    return result, nil
}
```

### Fluxo de Uso

```bash
# First time setup
kvstok export --encrypt
# Prompt: "Enter passphrase to encrypt export: "
# Saved: ~/.config/kvstok/kvstok.json (criptografado)
#        ~/.config/kvstok/kvstok.salt (salt do Argon2)

# On import
kvstok import
# Prompt: "Enter passphrase: "
# Auto-decrypta e importa
```

### Esforço: 4h
### Benefício: Crítico - Segurança necessária

---

## Feature #4: Multi-Device Sync (P2P)

### Descrição
Sincronizar keys entre dispositivos via end-to-end encrypted tunnel (sem servidor).

### Especificação Técnica

```bash
# Device A (Origin)
kvstok pair --device laptop
# Output: "Pairing code: 847392"

# Device B (Target)
kvstok pair --code 847392
# Sincroniza via:
# 1. mDNS discovery (local network)
# 2. DTLS over UDP (encrypted P2P)
# 3. Se não conseguir local, usa Signal Protocol relay (sem servidor)

# Verificar sincronização
kvstok sync status
# Dispositivos sincronizados: laptop (last sync: 2m ago)
```

### Implementação

```go
// internal/sync/peer_discovery.go - NOVO

type PeerManager struct {
    deviceName string
    publicKey  *rsa.PublicKey
    mdnsServer *zeroconf.Server
}

func (p *PeerManager) DiscoverPeers() []Peer {
    // Usar mDNS para descobrir outros kvstok peers na rede local
}

// internal/sync/dtls.go - NOVO
func (p *PeerManager) SyncWithPeer(peer Peer) error {
    // Estabelecer DTLS connection
    // Transferir delta de keys apenas
    // Usar Signal Protocol para E2E encryption
}
```

### Dependências
```go
// go.mod
require (
    github.com/grandcat/zeroconf v1.0.1 // mDNS
    github.com/pion/dtls/v2 v2.2.0      // DTLS
    github.com/signal-golang/signal-protocol-go v0.x.x  // E2E
)
```

### Esforço: 8h
### Benefício: Muito Alto - Multi-device access

---

## Feature #5: Web UI Dashboard

### Descrição
Dashboard web opcional para CRUD remoto seguro (sem replicar para cloud).

### Especificação Técnica

```bash
# Iniciar servidor web
kvstok web --port 8080 --auth basicauth
# Abre http://localhost:8080
# Login: admin/[password aleatório gerado]
```

### Stack
```go
// Backend: fiber (lightweight, fast)
// Frontend: htmx + Bulma CSS (simples, sem build step)
// Auth: JWT (gerado no inicio)

// main.go
import "github.com/gofiber/fiber/v3"

func startWebServer(port int) {
    app := fiber.New()
    
    // Middleware de auth
    app.Use(jwtMiddleware)
    
    // Routes
    app.Get("/api/keys", handleListKeys)
    app.Post("/api/keys", handleAddKey)
    app.Delete("/api/keys/:key", handleDeleteKey)
    app.Put("/api/keys/:key", handleUpdateKey)
    
    // Servir HTML estático
    app.Static("/", "./web/static")
    
    app.Listen(":" + string(rune(port)))
}
```

### UI (HTML + HTMX)
```html
<!-- web/static/index.html -->
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9/css/bulma.min.css">
    <script src="https://unpkg.com/htmx.org"></script>
</head>
<body>
    <div class="container">
        <h1>KVStoK Dashboard</h1>
        
        <form hx-post="/api/keys" hx-target="#keys-list">
            <input name="key" placeholder="Key" required>
            <input name="value" type="password" placeholder="Value" required>
            <button type="submit">Add</button>
        </form>
        
        <div id="keys-list" hx-get="/api/keys" hx-trigger="load, keyAdded">
            <!-- Populated by HTMX -->
        </div>
    </div>
</body>
</html>
```

### Esforço: 6h
### Benefício: Alto - Acesso remoto seguro

---

## Feature #6: CI/CD Integration

### Descrição
Integração com GitHub Actions, GitLab CI, e outros pipelines.

### GitHub Actions Example

```yaml
# .github/workflows/deploy.yml
name: Deploy with KVStoK

on: [push]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: waldirborbajr/kvstok@v1
        with:
          db-path: ${{ secrets.KVSTOK_DB_PATH }}
          export-format: dotenv
          output-file: .env
      
      - run: |
          source .env
          # Deploy usando variáveis do KVStoK
          docker push $DOCKER_REGISTRY/$IMAGE_NAME
```

### Implementação CLI

```go
// cmd/commands/env.go - NOVO

var EnvCmd = &cobra.Command{
    Use:     "env",
    Short:   "Export keys as environment variables.",
    Flags: map[string]string{
        "--format": "dotenv|shell|json (default: dotenv)",
    },
    Run: func(cmd *cobra.Command, args []string) {
        format, _ := cmd.Flags().GetString("format")
        
        database.DB.View(func(tx *nutsdb.Tx) error {
            keys, values, _ := tx.GetAll(database.Bucket)
            
            for i := 0; i < len(keys); i++ {
                k := string(keys[i])
                v := string(values[i])
                
                switch format {
                case "dotenv":
                    fmt.Printf("export %s='%s'\n", k, escapeShell(v))
                case "json":
                    // JSON output
                case "shell":
                    // Shell export
                }
            }
            return nil
        })
    },
}

func escapeShell(s string) string {
    return strings.ReplaceAll(s, "'", "'\\''")
}
```

### Esforço: 2h
### Benefício: Médio-Alto - DevOps workflow

---

## Feature #7: Audit Log

### Descrição
Registrar todas as operações (add, get, delete, export) com timestamp e contexto.

### Especificação Técnica

```go
// internal/audit/logger.go - NOVO

type AuditEntry struct {
    Timestamp time.Time
    Operation string // add, get, del, exp, imp
    Key       string
    User      string     // $USER env var
    Success   bool
    ErrorMsg  string
}

func LogOperation(op string, key string, success bool, err string) {
    entry := AuditEntry{
        Timestamp: time.Now(),
        Operation: op,
        Key:       key,
        User:      os.Getenv("USER"),
        Success:   success,
        ErrorMsg:  err,
    }
    
    // Salvar em ~/.config/kvstok/audit.log
    // Formato: JSON lines para easy parsing
}
```

### Visualização

```bash
kvstok audit log
# 2026-05-21 10:30:15 | get  | aws_token    | user | ✓
# 2026-05-21 10:30:20 | add  | new_key      | user | ✓
# 2026-05-21 10:30:25 | del  | old_key      | user | ✓

kvstok audit log --since "1 hour ago"
kvstok audit log --user root
kvstok audit log --operation get
```

### Esforço: 2h
### Benefício: Médio - Segurança + compliance

---

## Feature #8: Passphrase for Private Key

### Descrição
Proteger `kvstok.priv` com passphrase (derivado com Argon2).

### Especificação Técnica

```go
// internal/security/key_protection.go - NOVO

func ProtectPrivateKey(privateKey *rsa.PrivateKey, passphrase string) []byte {
    // 1. Gerar salt aleatório
    salt := make([]byte, 16)
    rand.Read(salt)
    
    // 2. Derivar key usando Argon2id
    key := argon2.IDKey(
        []byte(passphrase),
        salt,
        3, 64*1024, 4, 32,
    )
    
    // 3. Encriptar private key com AES-256-GCM
    privPEM := x509.MarshalPKCS1PrivateKey(privateKey)
    
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    
    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)
    
    ciphertext := gcm.Seal(nonce, nonce, privPEM, nil)
    
    // 4. Combinar salt + nonce + ciphertext
    return append(salt, append(nonce, ciphertext...)...)
}

func UnprotectPrivateKey(protected []byte, passphrase string) *rsa.PrivateKey {
    // Reverter processo acima
}
```

### Fluxo

```bash
# First run
kvstok init
# Enter passphrase to protect private key: [input]
# Saved: ~/.config/kvstok/kvstok.priv (encrypted)

# On startup
# kvstok getkv mykey
# Enter passphrase: [input]
# myvalue
```

### Esforço: 3h
### Benefício: Alto - Segurança necessária

---

## Feature #9: Key Rotation

### Descrição
Suportar versionamento de keys com rotação automática.

### Exemplos

```bash
# Adicionar versão a uma key
kvstok add api_token token_v1 --version
# Salva como: api_token@2026-05-21T10:30:00Z

# Listar histórico
kvstok history api_token
# api_token@2026-05-21T10:30:00Z
# api_token@2026-05-20T10:00:00Z
# api_token@2026-05-19T15:30:00Z

# Obter versão específica
kvstok get api_token@2026-05-20T10:00:00Z

# Rotacionar (deletar old versions)
kvstok rotate api_token --keep 3 --older-than "7d"
```

### Esforço: 3h
### Benefício: Médio - Compliance + security

---

## Feature #10: Shell Integration

### Descrição
Auto-complete e integração com shell rc files.

### Implementação

```bash
# Instalar auto-complete
kvstok completion bash | sudo tee /etc/bash_completion.d/kvstok
kvstok completion zsh  | sudo tee /usr/share/zsh/site-functions/_kvstok
kvstok completion fish | sudo tee /usr/share/fish/vendor_completions.d/kvstok.fish

# Adicionar alias ao ~/.bashrc
eval "$(kvstok shell alias)"
# Adiciona: alias kv=kvstok

# Depois use:
kv a mykey myvalue
kv g mykey
```

### Esforço: 1.5h
### Benefício: Médio - UX improvement

---

## 📊 Priorização Visual

```
CRÍTICA (Fazer ASAP):
├─ ✅ Fix Bugs (FIX #1-7) - 0.5h
├─ 🔒 Encrypt Exports (#3) - 4h
├─ 🔐 Passphrase Protection (#8) - 3h
└─ 🔍 Search/Filter (#1) - 2h
   Subtotal: ~9.5h

IMPORTANTE (Próximas 3-4 semanas):
├─ 🏷️  Tags/Categories (#2) - 3h
├─ 📡 Multi-Device Sync (#4) - 8h
├─ 📊 Web UI (#5) - 6h
└─ 📋 Audit Log (#7) - 2h
   Subtotal: ~19h

LEGAL (Nice to have):
├─ 🔄 CI/CD Integration (#6) - 2h
├─ 🔑 Key Rotation (#9) - 3h
└─ 🎯 Shell Integration (#10) - 1.5h
   Subtotal: ~6.5h
```

---

## 🎯 Roadmap Proposto

### Mês 1 (Junho 2026)
- Aplicar todos os bug fixes
- Implementar Encrypt Exports
- Adicionar Search/Filter
- Publicar v1.1

### Mês 2-3 (Julho-Agosto)
- Tags/Categories
- Multi-Device Sync (MVP)
- Audit Log
- Publicar v1.2

### Mês 4 (Setembro)
- Web UI (experimental)
- Key Rotation
- Shell Integration
- Publicar v1.3

### Mês 5+ (Outubro+)
- Melhorias based on user feedback
- Performance optimization
- Documentação completa

---

## 📝 Estimativa de Viabilidade

| Feature | Viabilidade | Complexidade | Impacto |
|---------|------------|-------------|--------|
| Search/Filter | ⭐⭐⭐⭐⭐ | 🟢 Baixa | 🟠 Alto |
| Tags | ⭐⭐⭐⭐⭐ | 🟡 Média | 🟠 Alto |
| Encrypt | ⭐⭐⭐⭐⭐ | 🟡 Média | 🔴 Crítico |
| P2P Sync | ⭐⭐⭐⭐ | 🔴 Alta | 🟠 Alto |
| Web UI | ⭐⭐⭐⭐ | 🔴 Alta | 🟠 Alto |
| CI/CD | ⭐⭐⭐⭐⭐ | 🟢 Baixa | 🟡 Médio |
| Audit | ⭐⭐⭐⭐⭐ | 🟢 Baixa | 🟡 Médio |
| Passphrase | ⭐⭐⭐⭐⭐ | 🟡 Média | 🔴 Crítico |
| Rotation | ⭐⭐⭐ | 🟡 Média | 🟡 Médio |
| Shell | ⭐⭐⭐⭐⭐ | 🟢 Baixa | 🟡 Médio |

---

**Recomendação:** Começar com Encrypt + Passphrase (segurança), depois Search/Tags (UX), depois P2P Sync (acesso remoto).
