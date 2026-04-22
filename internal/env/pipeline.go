package env

import (
	"fmt"
	"io"
	"net"
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

// GetJavaExec returns the path to the real java binary inside JDKDir
func (p *Pipeline) GetJavaExec() (string, error) {
	var javaExec string
	err := filepath.Walk(p.JDKDir, func(path string, info os.FileInfo, err error) error {
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
		return "", fmt.Errorf("java executable not found")
	}
	return javaExec, nil
}

// CheckJDK checks for isolation JDK and downloads it if missing
func (p *Pipeline) CheckJDK() error {
	javaExec, err := p.GetJavaExec()
	if err == nil {
		if _, err := os.Stat(javaExec); err == nil {
			ui.Success("JDK Isolado (Zero Install) encontrado.")
			return nil
		}
	}

	return ui.RunWithSpinner("Baixando e extraindo JRE 21 Isolado (Zero Install)...", func() error {
		osName := "linux"
		archName := "x64" // default
		ext := ".tar.gz"

		// Very basic OS detection for Adoptium URL format
		if os.PathSeparator == '\\' {
			osName = "windows"
			ext = ".zip"
		} else {
			// On unix check if it's macOS
			if _, err := os.Stat("/Applications"); err == nil {
				osName = "mac"
			}
		}

		adoptiumUrl := fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/21/ga/%s/%s/jre/hotspot/normal/eclipse", osName, archName)
		
		tempArchive := filepath.Join(p.BaseDir, "jre-temp"+ext)
		
		if err := DownloadFile(adoptiumUrl, tempArchive); err != nil {
			return fmt.Errorf("falha ao baixar JRE: %v", err)
		}
		defer os.Remove(tempArchive)

		if err := os.MkdirAll(p.JDKDir, 0755); err != nil {
			return err
		}

		if err := Extract(tempArchive, p.JDKDir); err != nil {
			return fmt.Errorf("falha ao extrair JRE: %v", err)
		}

		// Change perms of java to executable
		javaExec, err = p.GetJavaExec()
		if err == nil {
			os.Chmod(javaExec, 0755)
		}
		return nil
	})
}

// CheckArtifact downloads simulador.jar from GitHub Releases if missing
func (p *Pipeline) CheckArtifact() error {
	if _, err := os.Stat(p.JarPath); err == nil {
		ui.Success("Simulador (hubsaude-simulador.jar) encontrado.")
		return nil
	}

	return ui.RunWithSpinner("Baixando Simulador Oficial do GitHub Releases...", func() error {
		url := "https://github.com/kyriosdata/runner/releases/download/v0.0.2/hubsaude-simulador-0.0.0-SNAPSHOT.jar"
		if err := DownloadFile(url, p.JarPath); err != nil {
			return fmt.Errorf("erro ao baixar jar: %v", err)
		}
		return nil
	})
}

// CheckPort verifies if the port is open and available
func (p *Pipeline) CheckPort() error {
	address := fmt.Sprintf(":%d", p.TargetPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("a porta %d já está em uso por outro aplicativo", p.TargetPort)
	}
	listener.Close()
	ui.Success(fmt.Sprintf("Porta %d está livre para uso.", p.TargetPort))
	return nil
}
