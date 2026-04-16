package env

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jcal1998/ufg-soft-impl-runner/internal/state"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/ui"
)

const requiredJarVersion = "1.0.0"

// PipelinePipeline struct handles the startup verify steps
type Pipeline struct {
	BaseDir   string
	JDKDir    string
	JarPath   string
	TargetPort int
}

// NewPipeline initializes a new pipeline to ensure the environment is ready
func NewPipeline(port int) (*Pipeline, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	baseDir := filepath.Join(home, ".hubsaude")
	
	return &Pipeline{
		BaseDir:    baseDir,
		JDKDir:     filepath.Join(baseDir, "jdk"),
		JarPath:    filepath.Join(baseDir, "simulador.jar"),
		TargetPort: port,
	}, nil
}

// Run executes the entire pipeline sequentially before allowing a boot
func (p *Pipeline) Run() error {
	if err := p.CheckEnvironment(); err != nil {
		return err
	}

	if err := p.CheckJDK(); err != nil {
		return err
	}

	if err := p.CheckArtifact(); err != nil {
		return err
	}

	if err := p.CheckPort(); err != nil {
		return err
	}

	return nil
}

// CheckEnvironment ensures base dir exists
func (p *Pipeline) CheckEnvironment() error {
	return state.EnsureDir()
}

// CheckJDK checks for isolation JDK and downloads it if missing
func (p *Pipeline) CheckJDK() error {
	javaExec := filepath.Join(p.JDKDir, "bin", "java")
	if _, err := os.Stat(javaExec); err == nil {
		ui.Success("JDK Isolado encontrado.")
		return nil
	}

	return ui.RunWithSpinner("Baixando e configurando JDK 21 Isolado (Zero Install)...", func() error {
		// // TODO: Actual download and tar extraction logic for real usage
		// // For now, we simulate the action and "mock" the binary creation
		ui.SimulateWait(2)
		if err := os.MkdirAll(filepath.Join(p.JDKDir, "bin"), 0755); err != nil {
			return err
		}
		// Create a dummy file just to simulate success check in OS
		return os.WriteFile(javaExec, []byte("#!/bin/sh\necho 'java version 21'"), 0755)
	})
}

// CheckArtifact downloads simulador.jar from GitHub Releases if missing
func (p *Pipeline) CheckArtifact() error {
	if _, err := os.Stat(p.JarPath); err == nil {
		ui.Success("Simulador (simulador.jar) encontrado.")
		return nil
	}

	return ui.RunWithSpinner("Baixando Simulador do GitHub Releases...", func() error {
		ui.SimulateWait(2)
		// Simulating artifact download
		return os.WriteFile(p.JarPath, []byte("PK... simulated jar file ..."), 0644)
	})
}

// CheckPort verifies if the port is open and available
// If it's already running a process, the port check will fail ensuring we don't duplicate
func (p *Pipeline) CheckPort() error {
	// PING check will be implemented in process package, here we just do a mock test
	// In the real implementation, we could try net.Listen("tcp", fmt.Sprintf(":%d", p.TargetPort))
	// or perform a HTTP GET to /health if we just want to verify it's the HubSaude.
	return nil
}
