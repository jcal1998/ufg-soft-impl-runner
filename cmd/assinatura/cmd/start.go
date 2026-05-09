package cmd

import (
	"fmt"
	"os"

	"github.com/jcal1998/ufg-soft-impl-runner/internal/env"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/process"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/state"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/ui"
	"github.com/spf13/cobra"
)

var port int

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o motor criptográfico (assinador.jar) em background",
	Run: func(cmd *cobra.Command, args []string) {
		st, err := state.Load()
		if err != nil {
			ui.Error(fmt.Sprintf("Falha ao ler state.json: %v", err))
			return
		}

		if process.IsAssinadorRunningCheck(st.Assinador) {
			ui.Error(fmt.Sprintf("O motor já está em execução no PID %d. Use 'assinatura stop' ou 'status'.", st.Assinador.PID))
			return
		}

		ui.Info("Verificando artefato assinador.jar...")
		jarPath := env.ResolveAssinadorPath()
		if jarPath == "" || !fileExists(jarPath) {
			ui.Error("Artefato assinador.jar não encontrado. Certifique-se de compilá-lo (mvn clean package) ou baixe a versão de produção.")
			return
		}
		ui.Success(fmt.Sprintf("Artefato encontrado em: %s", jarPath))

		ui.Info("Iniciando o motor em background...")

		err = process.StartAssinadorBackground(port, jarPath)
		if err != nil {
			ui.Error(fmt.Sprintf("Não foi possível iniciar o motor: %v", err))
			return
		}

		ui.Success(fmt.Sprintf("Motor inicializado com sucesso na porta :%d", port))
	},
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVarP(&port, "port", "p", 8081, "Porta local para o motor criptográfico")
}
