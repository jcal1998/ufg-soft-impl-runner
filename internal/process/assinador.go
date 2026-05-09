package process

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jcal1998/ufg-soft-impl-runner/internal/state"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/ui"
)

// StartAssinadorBackground executa o Assinador Java em background e monitoriza a saúde até dar boot
func StartAssinadorBackground(port int, jarPath string) error {
	home, _ := os.UserHomeDir()
	jdkDir := filepath.Join(home, ".hubsaude", "jdk")
	jdkPath, err := findJavaExec(jdkDir)
	if err != nil {
		return fmt.Errorf("JDK não encontrado: %v. Execute o simulador start primeiro para baixar o JDK", err)
	}

	cmd := exec.Command(jdkPath, "-jar", jarPath, fmt.Sprintf("--server.port=%d", port))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("falha ao iniciar o processo Java do Assinador: %v", err)
	}

	// Update state
	currentState, _ := state.Load()
	currentState.Assinador = &state.AssinadorState{
		PID:         cmd.Process.Pid,
		Port:        port,
		JDKPath:     jdkPath,
		JarPath:     jarPath,
		LastStarted: time.Now(),
	}
	state.Save(currentState)

	return WaitForAssinadorHealth(port)
}

// WaitForAssinadorHealth aguarda ativamente a API retornar resposta HTTP (Assinador usa HTTP puro no momento)
func WaitForAssinadorHealth(port int) error {
	url := fmt.Sprintf("http://localhost:%d/health", port)
	client := http.Client{Timeout: 2 * time.Second}

	return ui.RunWithSpinner("Aguardando inicialização do Motor Criptográfico...", func() error {
		for i := 0; i < 40; i++ { // wait up to 20 seconds
			resp, err := client.Get(url)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					return nil
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
		return fmt.Errorf("timeout longo: O motor não respondeu na porta %d", port)
	})
}

// IsAssinadorRunningCheck verifica se existe um processo ativo e respondendo na web
func IsAssinadorRunningCheck(st *state.AssinadorState) bool {
	if st == nil || st.PID == 0 {
		return false
	}
	// Verify if process still exists in OS
	process, err := os.FindProcess(st.PID)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0)) // ping alive
	if err != nil {
		return false
	}

	// Double check with HTTP
	url := fmt.Sprintf("http://localhost:%d/health", st.Port)
	client := http.Client{Timeout: 1 * time.Second}

	resp, err := client.Get(url)
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == 200 {
			return true
		}
	}
	return false
}

// StopAssinador encerra o assinador de forma controlada
func StopAssinador(st *state.AssinadorState) error {
	if !IsAssinadorRunningCheck(st) {
		return fmt.Errorf("processo não está em execução ou não responde")
	}

	// Como o Assinador é simples e usa Javalin sem rota /shutdown construída explicitamente ainda,
	// podemos usar SIGTERM de forma segura, já que ele não lida com banco de dados.
	process, err := os.FindProcess(st.PID)
	if err == nil {
		process.Signal(syscall.SIGTERM)
	}

	ui.Success(fmt.Sprintf("Processo do Assinador (PID: %d) finalizado.", st.PID))

	// Clean up state
	currentState, _ := state.Load()
	currentState.Assinador = nil
	state.Save(currentState)

	return nil
}
