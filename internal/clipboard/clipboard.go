// internal/clipboard/clipboard.go
package clipboard

import (
	"fmt"

	"golang.design/x/clipboard"
)

// Copy copia texto para o clipboard
func Copy(text string) error {
	if text == "" {
		return fmt.Errorf("nada para copiar")
	}

	// Inicializa o clipboard (necessário apenas uma vez)
	clipboard.Init()

	err := clipboard.Write(clipboard.FmtText, []byte(text))
	if err != nil {
		return fmt.Errorf("falha ao copiar para o clipboard: %w", err)
	}

	return nil
}

// CopyWithConfirmation copia e mostra mensagem amigável
func CopyWithConfirmation(text, key string) error {
	if err := Copy(text); err != nil {
		return err
	}

	fmt.Printf("✅ Valor da chave '%s' copiado para o clipboard!\n", key)
	return nil
}
