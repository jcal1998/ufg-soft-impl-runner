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
	Short: "Interrompe a execução do motor criptográfico",
	Run: func(cmd *cobra.Command, args []string) {
		st, err := state.Load()
		if err != nil {
			ui.Error(fmt.Sprintf("Falha ao ler state.json: %v", err))
			return
		}

		if st.Assinador == nil || st.Assinador.PID == 0 {
			ui.Info("Não há nenhum motor registrado para parar.")
			return
		}

		ui.Info("Finalizando o motor...")
		err = process.StopAssinador(st.Assinador)
		if err != nil {
			ui.Error(fmt.Sprintf("Falha ao parar o processo: %v", err))
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
