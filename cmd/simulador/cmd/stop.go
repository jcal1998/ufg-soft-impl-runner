package cmd

import (
	"fmt"

	"github.com/jcal1998/ufg-soft-impl-runner/internal/process"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/state"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/ui"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Interrompe o simulador caso esteja em execução",
	Run: func(cmd *cobra.Command, args []string) {
		st, err := state.Load()
		if err != nil {
			ui.Error(fmt.Sprintf("Falha ao ler state.json: %v", err))
			return
		}

		if st.Simulador == nil {
			ui.Info("Não há nenhum simulador rodando para interromper.")
			return
		}

		err = process.Stop(st.Simulador)
		if err != nil {
			ui.Error(fmt.Sprintf("Erro ao matar o processo: %v", err))
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
