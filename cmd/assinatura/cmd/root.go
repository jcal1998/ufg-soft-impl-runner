package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "assinatura",
	Short: "CLI para assinatura de pacotes FHIR",
	Long:  `O CLI assinatura permite iniciar o motor criptográfico em Java e assinar documentos JSON seguindo os padrões da SES-GO.`,
}

// Execute executa o comando raiz.
func Execute() error {
	return rootCmd.Execute()
}
