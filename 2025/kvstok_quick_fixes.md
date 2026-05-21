# 🔧 KVSTOK - QUICK FIX GUIDE

## Patches para Aplicar Imediatamente (30-45 min)

---

## ✅ FIX #1: Corrigir Export Command (Dados Vazios)

**Arquivo:** `cmd/commands/export.go`

**Problema:** A variável `content` nunca é preenchida com dados antes de ser exportada.

**Localizar (linhas ~28-37):**
```go
if keys, values, err := tx.GetAll(database.Bucket); err != nil {
    return err
} else {
    n := len(keys)
    for i := 0; i < n; i++ {
        fmt.Println(string(keys[i]), " ", string(values[i]))
    }
}

configFile := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.json"
```

**Substituir por:**
```go
if keys, values, err := tx.GetAll(database.Bucket); err != nil {
    return err
} else {
    n := len(keys)
    for i := 0; i < n; i++ {
        // ✅ ADICIONAR ESTA LINHA - Preencher o mapa!
        content[string(keys[i])] = string(values[i])
        fmt.Println(string(keys[i]), " ", string(values[i]))
    }
}

configFile := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.json"
```

**Comando git (single line):**
```bash
git apply << 'EOF'
--- a/cmd/commands/export.go
+++ b/cmd/commands/export.go
@@ -33,6 +33,7 @@ var ExpCmd = &cobra.Command{
 				} else {
 					n := len(keys)
 					for i := 0; i < n; i++ {
+						content[string(keys[i])] = string(values[i])
 						fmt.Println(string(keys[i]), " ", string(values[i]))
 					}
 				}
EOF
```

---

## ✅ FIX #2: Corrigir Get Command (Performance Bug)

**Arquivo:** `cmd/commands/get.go`

**Problema:** Usa `Update()` em vez de `View()` para operação de leitura = lock desnecessário

**Localizar (linhas ~18-27):**
```go
Run: func(cmd *cobra.Command, args []string) {
    // nolint:staticcheck
    if err := database.DB.Update(
        func(tx *nutsdb.Tx) error {
            key := []byte(args[0])
            content, err := tx.Get(database.Bucket, key)
            must.Must(err, "GetCmd() - key not found or datababse must be empty.")
            fmt.Printf("%s\n", content)
            return nil
        }); err != nil {
    }
},
```

**Substituir por:**
```go
Run: func(cmd *cobra.Command, args []string) {
    // nolint:staticcheck
    if err := database.DB.View(  // ✅ Mudar Update para View
        func(tx *nutsdb.Tx) error {
            key := []byte(args[0])
            content, err := tx.Get(database.Bucket, key)
            must.Must(err, "GetCmd() - key not found or database must be empty.")
            fmt.Printf("%s\n", content)
            return nil
        }); err != nil {
        must.Must(err, "GetCmd() - failed to retrieve key")  // ✅ Tratar erro
    }
},
```

**Patch completo:**
```diff
--- a/cmd/commands/get.go
+++ b/cmd/commands/get.go
@@ -17,15 +17,16 @@ var GetCmd = &cobra.Command{
 	Run: func(cmd *cobra.Command, args []string) {
 		// nolint:staticcheck
-		if err := database.DB.Update(
+		if err := database.DB.View(
 			func(tx *nutsdb.Tx) error {
 				key := []byte(args[0])
 				content, err := tx.Get(database.Bucket, key)
-				must.Must(err, "GetCmd() - key not found or datababse must be empty.")
+				must.Must(err, "GetCmd() - key not found or database must be empty.")
 				fmt.Printf("%s\n", content)
 				return nil
 			}); err != nil {
+			must.Must(err, "GetCmd() - failed to retrieve key")
 		}
 	},
 }
```

---

## ✅ FIX #3: Renomear `secutiry` → `security`

**Problema:** Typo em nome de package crítico causa confusão

### Passo 1: Renomear pasta
```bash
git mv internal/secutiry internal/security
```

### Passo 2: Atualizar import em `main.go` (linha ~8)

**Localizar:**
```go
import (
    // ...
    security "github.com/waldirborbajr/kvstok/internal/secutiry"
)
```

**Substituir por:**
```go
import (
    // ...
    security "github.com/waldirborbajr/kvstok/internal/security"
)
```

**Patch:**
```diff
--- a/main.go
+++ b/main.go
@@ -6,7 +6,7 @@ import (
 
 	"github.com/waldirborbajr/kvstok/cmd"
 	"github.com/waldirborbajr/kvstok/internal/kvpath"
-	security "github.com/waldirborbajr/kvstok/internal/secutiry"
+	security "github.com/waldirborbajr/kvstok/internal/security"
 )
```

---

## ✅ FIX #4: Remover Dead Code

**Arquivo:** `cmd/root.go`

**Problema:** Constante `DBSIZE` nunca é usada

**Localizar (linha ~6):**
```go
// Size of database to store key/value
const DBSIZE = 2048 * 2048
```

**Ação:** Remover linhas 5-6 completamente

**Patch:**
```diff
--- a/cmd/root.go
+++ b/cmd/root.go
@@ -3,9 +3,6 @@ package cmd
 import (
 	"log"
 
 	"github.com/nutsdb/nutsdb"
 	"github.com/spf13/cobra"
 	"github.com/waldirborbajr/kvstok/cmd/commands"
@@ -13,9 +10,6 @@ import (
 	"github.com/waldirborbajr/kvstok/internal/kvpath"
 	"github.com/waldirborbajr/kvstok/internal/must"
 	"github.com/waldirborbajr/kvstok/internal/version"
-)

-// Size of database to store key/value
-const DBSIZE = 2048 * 2048

 // rootCmd represents the base command when called without any subcommands
```

---

## ✅ FIX #5: Simplificar `isEquals` Function

**Arquivo:** `cmd/commands/import.go`

**Problema:** Função desnecessária, pode ser substituída por comparação direta

**Localizar (linhas ~42-43 e 62-68):**

**Linha 42-43 (remover):**
```go
areEquals := isEquals(currentHash, string(storedHash))
```

**Substituir por:**
```go
areEquals := currentHash == string(storedHash)
```

**Remover função inteira (linhas 62-68):**
```go
func isEquals(param1 string, param2 string) bool {
    bret := true
    if param1 != param2 {
        bret = false
    }
    return bret
}
```

**Patch completo:**
```diff
--- a/cmd/commands/import.go
+++ b/cmd/commands/import.go
@@ -38,7 +38,7 @@ var ImpCmd = &cobra.Command{
 		currentHash := kvpath.GenHash(configFile)
 		storedHash := []byte(file)
 
-		areEquals := isEquals(currentHash, string(storedHash))
+		areEquals := currentHash == string(storedHash)
 
 		if !areEquals {
 
@@ -57,13 +57,0 @@ var ImpCmd = &cobra.Command{
 		fmt.Printf("Keys imported successfully.")
 	},
 }
-
-func isEquals(param1 string, param2 string) bool {
-	bret := true
-
-	if param1 != param2 {
-		bret = false
-	}
-
-	return bret
-}
```

---

## ✅ FIX #6: Typo em Mensagem de Erro

**Arquivo:** `cmd/commands/get.go`

**Problema:** "datababse" deveria ser "database"

**Localizar (linha ~23):**
```go
must.Must(err, "GetCmd() - key not found or datababse must be empty.")
```

**Substituir por:**
```go
must.Must(err, "GetCmd() - key not found or database must be empty.")
```

---

## ✅ FIX #7: Tratar Erro em Import

**Arquivo:** `cmd/commands/import.go`

**Problema:** `json.Unmarshal` não verifica erro (linha ~50)

**Localizar (linhas ~48-50):**
```go
file, err = os.ReadFile(configFile)
must.Must(err, "ImpCmd() - oops! Huston, we have a problem integrity broken.")

json.Unmarshal([]byte(file), &dataResult)
```

**Substituir por:**
```go
file, err = os.ReadFile(configFile)
must.Must(err, "ImpCmd() - oops! Huston, we have a problem integrity broken.")

err = json.Unmarshal([]byte(file), &dataResult)  // ✅ Capturar erro
must.Must(err, "ImpCmd() - failed to parse JSON file")
```

**Patch:**
```diff
--- a/cmd/commands/import.go
+++ b/cmd/commands/import.go
@@ -46,7 +46,8 @@ var ImpCmd = &cobra.Command{
 		if areEquals {
 			// Import JSON after integrity check
 			file, err = os.ReadFile(configFile)
 			must.Must(err, "ImpCmd() - oops! Huston, we have a problem integrity broken.")
 
-			json.Unmarshal([]byte(file), &dataResult)
+			err = json.Unmarshal([]byte(file), &dataResult)
+			must.Must(err, "ImpCmd() - failed to parse JSON file")
 
 			for key, value := range dataResult {
```

---

## 🎯 ORDEM DE APLICAÇÃO RECOMENDADA

```bash
# 1. Copiar branch de desenvolvimento
git checkout -b fix/critical-bugs

# 2. Aplicar patches na ordem (testes após cada um)
# FIX #1 - Export
# FIX #2 - Get Command
# FIX #3 - Renomear security
# FIX #4 - Dead code
# FIX #5 - isEquals
# FIX #6 - Typos
# FIX #7 - JSON error

# 3. Testes
go test ./...

# 4. Build
go build -o kvstok

# 5. Teste manual
./kvstok a test value
./kvstok e  # Deve gerar JSON com dados
./kvstok l
./kvstok g test

# 6. Commit
git commit -m "fix: resolve critical bugs in get, export, and security import"
```

---

## 🧪 Teste Cada Fix

### Depois do FIX #1 (Export):
```bash
kvstok a key1 value1
kvstok a key2 value2
kvstok e
cat ~/.config/kvstok/kvstok.json
# Deve mostrar: { "key1": "value1", "key2": "value2" }
```

### Depois do FIX #2 (Get):
```bash
kvstok a testkey secretvalue
kvstok g testkey
# Deve retornar: secretvalue (sem erro)
```

### Depois do FIX #3 (Security rename):
```bash
go build -o kvstok
# Sem erro de import
```

---

## 📈 Próximos Passos Após Patches

1. ✅ Adicionar testes unitários básicos (4h)
2. ✅ Adicionar linting (golangci-lint)
3. ✅ Documentar mudanças no CHANGELOG
4. ✅ Release v1.1 com bug fixes
5. ✅ Começar FASE 2: Encriptação

---

**Tempo total de aplicação:** ~45 minutos  
**Benefício:** Corrigir 7 bugs críticos que afetam funcionalidade central
