package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// SimuladorState representa o estado atual do simulador.jar rodando.
type SimuladorState struct {
	PID         int       `json:"pid"`
	Port        int       `json:"port"`
	JDKPath     string    `json:"jdk_path"`
	JarVersion  string    `json:"jar_version"`
	LastStarted time.Time `json:"last_started"`
}

// HubSaudeState representa a raiz do state.json
type HubSaudeState struct {
	Simulador *SimuladorState `json:"simulador,omitempty"`
}

var stateFilePath string

func init() {
	home, err := os.UserHomeDir()
	if err == nil {
		stateFilePath = filepath.Join(home, ".hubsaude", "state.json")
	}
}

// EnsureDir cria o diretório ~/.hubsaude caso não exista.
func EnsureDir() error {
	dir := filepath.Dir(stateFilePath)
	return os.MkdirAll(dir, 0755)
}

// Load lê o state.json e retorna o estado populado. Retorna um estado vazio se não existir.
func Load() (*HubSaudeState, error) {
	if _, err := os.Stat(stateFilePath); os.IsNotExist(err) {
		return &HubSaudeState{}, nil
	}

	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		return nil, err
	}

	var state HubSaudeState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// Save grava o estado atual no state.json.
func Save(state *HubSaudeState) error {
	if err := EnsureDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFilePath, data, 0644)
}
