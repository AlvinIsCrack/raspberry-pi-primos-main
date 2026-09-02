package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	_url "net/url"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

func computeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type UpdateResult struct {
	Updated        bool   `json:"updated"`
	CurrentVersion string `json:"current_version"`
	TargetVersion  string `json:"target_version"`
	Message        string `json:"message"`
}

type UpdaterService struct {
	repoOwner string
	repoName  string
	client    *http.Client
	mu        sync.Mutex
}

func NewUpdaterService(owner, repo string) *UpdaterService {
	return &UpdaterService{
		repoOwner: owner,
		repoName:  repo,
		client:    &http.Client{Timeout: 45 * time.Second},
	}
}

// GetRunningVersion obtiene la versión embebida por Go en el binario (o dev/unknown si no se fijó).
func GetRunningVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func (s *UpdaterService) CheckAndApply(targetVersion string) (*UpdateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentVer := GetRunningVersion()

	// Si se pide una versión explícita y ya la tenemos en ejecución, no hacemos nada
	if targetVersion != "" && targetVersion == currentVer {
		return &UpdateResult{
			Updated:        false,
			CurrentVersion: currentVer,
			TargetVersion:  targetVersion,
			Message:        "El sistema ya se encuentra en la versión solicitada",
		}, nil
	}

	// Si no se especifica, busca la más reciente (/latest); sino busca el tag concreto (/tags/{targetVersion})
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", s.repoOwner, s.repoName)
	if targetVersion != "" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s",
			s.repoOwner, s.repoName, _url.PathEscape(targetVersion))
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fallo al consultar GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api devolvió status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("error decodificando release: %w", err)
	}

	// Comprobación si la release encontrada es idéntica a la que se está corriendo
	if release.TagName == "" || (currentVer != "dev" && release.TagName == currentVer) {
		return &UpdateResult{
			Updated:        false,
			CurrentVersion: currentVer,
			TargetVersion:  release.TagName,
			Message:        "El sistema ya se encuentra en la versión disponible",
		}, nil
	}

	var downloadURL string
	for _, asset := range release.Assets {
		parts := strings.Split(strings.TrimSuffix(asset.Name, filepath.Ext(asset.Name)), "_")
		if len(parts) >= 2 && parts[len(parts)-2] == runtime.GOOS && parts[len(parts)-1] == runtime.GOARCH {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return nil, errors.New("no se encontró binario compatible en los assets de la release")
	}

	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo path del ejecutable: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return nil, fmt.Errorf("error resolviendo symlinks del ejecutable: %w", err)
	}

	tmpPath := exePath + ".tmp"
	if err := s.downloadFile(downloadURL, tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		return nil, fmt.Errorf("error descargando binario: %w", err)
	}

	currentHash, errCur := computeSHA256(exePath)
	if errCur != nil {
		_ = os.Remove(tmpPath)
		return nil, fmt.Errorf("error calculando checksum del binario actual: %w", errCur)
	}

	downloadedHash, errDown := computeSHA256(tmpPath)
	if errDown != nil {
		_ = os.Remove(tmpPath)
		return nil, fmt.Errorf("error calculando checksum del binario descargado: %w", errDown)
	}

	if currentHash == downloadedHash {
		_ = os.Remove(tmpPath)
		return &UpdateResult{
			Updated:        false,
			CurrentVersion: currentVer,
			TargetVersion:  release.TagName,
			Message:        "El binario descargado es idéntico al actual (checksum coincide)",
		}, nil
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		_ = os.Remove(tmpPath)
		return nil, fmt.Errorf("error aplicando permisos de ejecución: %w", err)
	}

	// Reemplazo atómico en Linux
	oldPath := exePath + ".old"
	_ = os.Remove(oldPath)

	if err := os.Rename(exePath, oldPath); err != nil {
		_ = os.Remove(tmpPath)
		return nil, fmt.Errorf("error al mover ejecutable actual: %w", err)
	}

	if err := os.Rename(tmpPath, exePath); err != nil {
		// Rollback
		_ = os.Rename(oldPath, exePath)
		_ = os.Remove(tmpPath)
		return nil, fmt.Errorf("error al aplicar nuevo binario: %w", err)
	}
	_ = os.Remove(oldPath)

	return &UpdateResult{
		Updated:        true,
		CurrentVersion: currentVer,
		TargetVersion:  release.TagName,
		Message:        "Binario actualizado exitosamente. Reiniciando servicio...",
	}, nil
}

func (s *UpdaterService) downloadFile(url, destPath string) error {
	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := s.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("descarga falló con status %d", resp.StatusCode)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}
