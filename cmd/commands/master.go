// cmd/commands/master.go
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var masterCmd = &cobra.Command{
	Use:   "master",
	Short: "Gerencia a senha mestra",
}

var masterStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Mostra o status da senha mestra",
	Run: func(cmd *cobra.Command, args []string) {
		store, _ := GetStore()
		defer store.Close()

		if store.sec.IsMasterPasswordSet() {
			fmt.Println("✅ Senha mestra configurada e ativa")
		} else {
			fmt.Println("❌ Senha mestra não configurada. Execute: kvstok init")
		}
	},
}

var masterChangeCmd = &cobra.Command{
	Use:   "change",
	Short: "Altera a senha mestra (em breve)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔄 Funcionalidade de troca de senha em desenvolvimento...")
		// Futuramente: re-criptografar todos os dados com nova senha
	},
}

func init() {
	masterCmd.AddCommand(masterStatusCmd)
	masterCmd.AddCommand(masterChangeCmd)
	rootCmd.AddCommand(masterCmd)
}
