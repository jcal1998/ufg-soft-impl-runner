package env

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type GitHubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []GitHubAsset `json:"assets"`
}

// GetLatestSimuladorRelease busca nas releases do GitHub a versão mais recente que contém o arquivo JAR do simulador
func GetLatestSimuladorRelease() (version string, downloadURL string, err error) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/kyriosdata/runner/releases")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("github api error: %s", resp.Status)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", "", err
	}

	for _, rel := range releases {
		for _, asset := range rel.Assets {
			// Busca o jar do simulador nas assets
			if strings.Contains(asset.Name, "simulador") && strings.HasSuffix(asset.Name, ".jar") {
				return rel.TagName, asset.BrowserDownloadURL, nil
			}
		}
	}

	return "", "", fmt.Errorf("nenhum simulador.jar encontrado nas releases do github")
}
