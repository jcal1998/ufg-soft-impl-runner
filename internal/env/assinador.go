package env

import (
	"os"
	"path/filepath"
)

// ResolveAssinadorPath executa a lógica de resolução de caminhos para encontrar o assinador.jar
func ResolveAssinadorPath() string {
	// 1. Variável de ambiente
	if path := os.Getenv("RUNNER_ASSINADOR_PATH"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 2. Caminho relativo (Desenvolvimento)
	relPath := filepath.Join("assinador-java", "target", "assinador.jar")
	if _, err := os.Stat(relPath); err == nil {
		if absPath, err := filepath.Abs(relPath); err == nil {
			return absPath
		}
	}

	// 3. Fallback de produção
	home, _ := os.UserHomeDir()
	prodPath := filepath.Join(home, ".hubsaude", "assinador.jar")
	return prodPath
}
