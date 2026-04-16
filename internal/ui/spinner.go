package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/huh/spinner"
)

// RunWithSpinner executa uma função enquanto exibe um spinner com animação
func RunWithSpinner(title string, action func() error) error {
	var actionErr error

	// huh/spinner block until action finishes
	err := spinner.New().
		Title(title).
		Action(func() {
			actionErr = action()
		}).
		Run()

	if err != nil {
		return err // erro do próprio spinner runner
	}

	return actionErr
}

// Success exibe uma mensagem de sucesso com cor verde nativa se aplicável ou texto normal
func Success(msg string) {
	fmt.Printf("✅ %s\n", msg)
}

// Info exibe uma mensagem informativa
func Info(msg string) {
	fmt.Printf("ℹ️  %s\n", msg)
}

// Error exibe uma mensagem de erro
func Error(msg string) {
	fmt.Printf("❌ %s\n", msg)
}

// SimulateWait é uma função helper para testes simulando o tempo de loading
func SimulateWait(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}
