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
	Short: "Verifica se o simulador está rodando",
	Run: func(cmd *cobra.Command, args []string) {
		st, err := state.Load()
		if err != nil {
			ui.Error(fmt.Sprintf("Não foi possível ler o estado: %v", err))
			return
		}

		if st.Simulador == nil {
			ui.Info("Nenhum simulador registrado. (Aberto/Parado)")
			return
		}

		if process.IsRunningCheck(st.Simulador) {
			ui.Success(fmt.Sprintf("O simulador ESTÁ NO AR. (PID: %d, Porta: %d, Versão: %s)", 
			st.Simulador.PID, st.Simulador.Port, st.Simulador.JarVersion))
		} else {
			ui.Error(fmt.Sprintf("O simulador (PID: %d) não responde. Pode ter morrido inesperadamente. Atualizando estado...", st.Simulador.PID))
			st.Simulador = nil
			state.Save(st) // clean ghost state
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
