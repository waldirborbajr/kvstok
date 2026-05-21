// cmd/commands/init.go
package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"golang.org/x/term"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa o kvstok configurando a senha mestra",
	Long: `Inicializa o kvstok criando a senha mestra necessária para criptografar todos os dados.

Esta senha será usada para proteger todos os seus segredos. Guarde-a com segurança!`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	store, err := database.NewStore("")
	if err != nil {
		return fmt.Errorf("erro ao inicializar banco: %w", err)
	}
	defer store.Close()

	// Verifica se já está inicializado
	if store.IsMasterPasswordSet() {
		fmt.Println("⚠️  kvstok já está inicializado.")
		fmt.Println("   Use 'kvstok master change' para alterar a senha mestra (futuro).")
		return nil
	}

	fmt.Println("🔐 Configuração da Senha Mestra do kvstok")
	fmt.Println("=========================================")
	fmt.Println("Esta senha será usada para proteger todos os seus dados.")
	fmt.Println("")

	password, err := readPassword("Digite a senha mestra: ")
	if err != nil {
		return err
	}

	if len(password) < 8 {
		return fmt.Errorf("a senha mestra deve ter no mínimo 8 caracteres")
	}

	confirm, err := readPassword("Confirme a senha mestra: ")
	if err != nil {
		return err
	}

	if password != confirm {
		return fmt.Errorf("as senhas não coincidem")
	}

	// Configura a senha mestra
	if err := store.SetMasterPassword(password); err != nil {
		return fmt.Errorf("falha ao configurar senha mestra: %w", err)
	}

	fmt.Println("\n✅ kvstok inicializado com sucesso!")
	fmt.Println("   Todos os dados agora serão criptografados.")
	fmt.Println("")
	fmt.Println("Dica: Você pode usar a flag --master para não digitar a senha toda vez:")
	fmt.Println("   kvstok --master SUASENHA add ...")

	return nil
}

// readPassword lê senha de forma segura (sem eco)
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Tenta usar terminal sem eco (melhor UX)
	if term.IsTerminal(int(syscall.Stdin)) {
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // nova linha
		return string(bytePassword), err
	}

	// Fallback para ambiente sem terminal (ex: scripts)
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	return strings.TrimSpace(password), err
}

// Helper para checar se master password já está configurada
func (s *Store) IsMasterPasswordSet() bool {
	return s.sec.IsMasterPasswordSet()
}
