package process

import (
	"fmt"
	"io"
	"net"
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
func StartBackground(port int) error {
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
		JarVersion:  "1.0.0", // mock versão
		LastStarted: time.Now(),
	}
	state.Save(currentState)

	return WaitForHealth(port)
}

// WaitForHealth aguarda ativamente a porta TCP ser aberta
func WaitForHealth(port int) error {
	address := fmt.Sprintf("localhost:%d", port)

	return ui.RunWithSpinner("Aguardando inicialização do Simulador (TCP Check)...", func() error {
		for i := 0; i < 60; i++ { // wait up to 30 seconds (500ms * 60) for heavy Spring Boot
			conn, err := net.DialTimeout("tcp", address, 1*time.Second)
			if err == nil {
				conn.Close()
				return nil // boot sucesso
			}
			time.Sleep(500 * time.Millisecond)
		}
		return fmt.Errorf("timeout longo: O simulador não abriu a porta %d", port)
	})
}

// IsRunningCheck verifica se existe um processo ativo e respondendo na porta TCP
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

	// Double check with TCP to ensure it's not a zombie PID taken by another app
	address := fmt.Sprintf("localhost:%d", st.Port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err == nil {
		conn.Close()
		return true
	}
	return false
}

// Stop mata o processo de forma limpa enviando SIGTERM
func Stop(st *state.SimuladorState) error {
	if !IsRunningCheck(st) {
		return fmt.Errorf("processo não está em execução")
	}

	process, err := os.FindProcess(st.PID)
	if err != nil {
		return err
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("falha ao enviar SIGTERM: %v", err)
	}

	ui.Success(fmt.Sprintf("Processo (PID: %d) finalizado graciosamente.", st.PID))

	// Clean up state
	currentState, _ := state.Load()
	currentState.Simulador = nil
	state.Save(currentState)

	return nil
}
