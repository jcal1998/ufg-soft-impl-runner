package process

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jcal1998/ufg-soft-impl-runner/internal/state"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/ui"
)

// findJavaExec returns the path to the real java binary inside JDKDir
func findJavaExec(jdkDir string) (string, error) {
	var javaExec string
	err := filepath.Walk(jdkDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (info.Name() == "java" || info.Name() == "java.exe") {
			javaExec = path
			return io.EOF // Stop walking once we find it
		}
		return nil
	})

	if err != nil && err != io.EOF {
		return "", err
	}
	if javaExec == "" {
		return "", fmt.Errorf("java executável não encontrado")
	}
	return javaExec, nil
}

// StartBackground executa o Java em background e monitoriza a saúde até dar boot
func StartBackground(port int, jarVersion string) error {
	home, _ := os.UserHomeDir()
	jdkDir := filepath.Join(home, ".hubsaude", "jdk")
	jdkPath, err := findJavaExec(jdkDir)
	if err != nil {
		return err
	}

	jarPath := filepath.Join(home, ".hubsaude", "simulador.jar")

	cmd := exec.Command(jdkPath, "-jar", jarPath, fmt.Sprintf("--server.port=%d", port))
	// Detach process from terminal session so it continues running
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("falha ao iniciar o processo Java: %v", err)
	}

	// Update state
	currentState, _ := state.Load()
	currentState.Simulador = &state.SimuladorState{
		PID:         cmd.Process.Pid,
		Port:        port,
		JDKPath:     jdkPath,
		JarVersion:  jarVersion, // Usa a versão dinâmica resolvida da API do Github
		LastStarted: time.Now(),
	}
	state.Save(currentState)

	return WaitForHealth(port)
}

// WaitForHealth aguarda ativamente a API retornar resposta HTTP (ignora SSL)
func WaitForHealth(port int) error {
	url := fmt.Sprintf("https://localhost:%d/api/info", port)

	// Como o certificado localhost é self-signed, precisamos pular a verificação TLS
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Timeout: 2 * time.Second, Transport: tr}

	return ui.RunWithSpinner("Aguardando inicialização do Simulador (HTTP Check)...", func() error {
		for i := 0; i < 40; i++ { // wait up to 20 seconds (500ms * 40)
			resp, err := client.Get(url)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					return nil
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
		return fmt.Errorf("timeout longo: O simulador não respondeu na porta %d", port)
	})
}

// IsRunningCheck verifica se existe um processo ativo e respondendo na web
func IsRunningCheck(st *state.SimuladorState) bool {
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

	// Double check with HTTP to ensure it's not a zombie PID taken by another app
	url := fmt.Sprintf("https://localhost:%d/api/info", st.Port)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Timeout: 1 * time.Second, Transport: tr}

	resp, err := client.Get(url)
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == 200 {
			return true
		}
	}
	return false
}

// Stop encerra o simulador de forma controlada via endpoint REST (graceful shutdown)
func Stop(st *state.SimuladorState) error {
	if !IsRunningCheck(st) {
		return fmt.Errorf("processo não está em execução ou não responde")
	}

	url := fmt.Sprintf("https://localhost:%d/shutdown", st.Port)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Timeout: 5 * time.Second, Transport: tr}

	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		// Fallback para SIGTERM caso o HTTP falhe
		ui.Error(fmt.Sprintf("Falha no graceful shutdown HTTP: %v. Forçando SIGTERM...", err))
		process, err := os.FindProcess(st.PID)
		if err == nil {
			process.Signal(syscall.SIGTERM)
		}
	} else {
		resp.Body.Close()
	}

	ui.Success(fmt.Sprintf("Processo (PID: %d) finalizado via /shutdown.", st.PID))

	// Clean up state
	currentState, _ := state.Load()
	currentState.Simulador = nil
	state.Save(currentState)

	return nil
}
