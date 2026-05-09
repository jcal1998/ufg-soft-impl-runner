package cmd

import (
	"fmt"

	"github.com/jcal1998/ufg-soft-impl-runner/internal/process"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/state"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exibe o status do motor criptográfico",
	Run: func(cmd *cobra.Command, args []string) {
		st, err := state.Load()
		if err != nil {
			ui.Error(fmt.Sprintf("Erro ao ler estado: %v", err))
			return
		}

		if process.IsAssinadorRunningCheck(st.Assinador) {
			ui.Success(fmt.Sprintf("O motor ESTÁ NO AR. (PID: %d, Porta: %d)", st.Assinador.PID, st.Assinador.Port))
		} else {
			ui.Info("Nenhum motor registrado ou em execução.")
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
