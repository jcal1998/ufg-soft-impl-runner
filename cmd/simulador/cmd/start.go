package cmd

import (
	"fmt"

	"github.com/jcal1998/ufg-soft-impl-runner/internal/env"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/process"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/state"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/ui"
	"github.com/spf13/cobra"
)

var port int

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o simulador do HubSaúde minimizando a complexidade no terminal",
	Run: func(cmd *cobra.Command, args []string) {
		// Verify absolute state
		st, err := state.Load()
		if err != nil {
			ui.Error(fmt.Sprintf("Falha ao ler state.json: %v", err))
			return
		}

		if process.IsRunningCheck(st.Simulador) {
			ui.Error(fmt.Sprintf("O simulador já está em execução no PID %d. Use 'simulador stop' ou 'status'.", st.Simulador.PID))
			return
		}

		ui.Info("Iniciando a fila de aprovisionamento (ACL)...")

		// Create validation pipeline
		pipeline, err := env.NewPipeline(port)
		if err != nil {
			ui.Error(fmt.Sprintf("Erro nas configurações básicas: %v", err))
			return
		}

		err = pipeline.Run()
		if err != nil {
			ui.Error(fmt.Sprintf("Falha na pipeline de verificação: %v", err))
			return
		}

		ui.Info("Ambiente pronto. Procedendo com arranque em background...")

		err = process.StartBackground(port)
		if err != nil {
			ui.Error(fmt.Sprintf("Não foi possível iniciar o servidor: %v", err))
			return
		}

		ui.Success(fmt.Sprintf("Simulador inicializado com sucesso na porta :%d", port))
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVarP(&port, "port", "p", 8080, "Porta local para o daemon do simulador")
}
