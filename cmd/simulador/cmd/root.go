package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "simulador",
	Short: "Gerenciador do Ciclo de Vida do Simulador do HubSaúde",
	Long:  `CLI multiplataforma focado exclusivamente em gerenciar o download, validação, e ciclo de vida do simulador.jar no background atuando como Anti-Corruption Layer.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Root flags
}
