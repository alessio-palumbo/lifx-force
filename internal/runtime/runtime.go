package runtime

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/alessio-palumbo/lifx-force/internal/config"
)

//go:embed assets/fingertrack.zip
var fingertrackZip []byte

// EnsureFingertrackInstalled makes sure the correct version is extracted.
func EnsureFingertrackInstalled(logger *slog.Logger) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	baseDir := filepath.Join(home, ".lifx-force")
	ftPath := filepath.Join(baseDir, "fingertrack")
	versionFile := filepath.Join(baseDir, ".version")

	// Check installed version matches current otherwise replace.
	if data, err := os.ReadFile(versionFile); err == nil {
		currentVersion := strings.TrimSpace(string(data))
		if currentVersion == FingertrackVersion {
			logger.Debug("Fingertrack version match requirements", slog.String("current_version", currentVersion))
			return ftPath, nil
		}
		logger.Debug("Fingertrack version does not match requirements",
			slog.String("current_version", currentVersion),
			slog.String("required_version", FingertrackVersion),
		)
	}

	logger.Info("Unzipping Fingertrack", slog.String("version", FingertrackVersion))

	if err := unzip(bytes.NewReader(fingertrackZip), int64(len(fingertrackZip)), baseDir); err != nil {
		return "", err
	}

	if err := os.WriteFile(versionFile, []byte(FingertrackVersion), 0644); err != nil {
		return "", err
	}

	return ftPath, nil
}

// ArgsFromConfig parses the given Config and returns a list of Fingertrack's arguments.
func ArgsFromConfig(cfg *config.Config) []string {
	return []string{
		"--frame-skip", fmt.Sprintf("%d", cfg.Tracking.FrameSkip),
		"--buffer-size", fmt.Sprintf("%d", cfg.Tracking.BufferSize),
		"--gesture-threshold", fmt.Sprintf("%.1f", cfg.Tracking.GestureThreshold),
	}
}

func unzip(data io.ReaderAt, size int64, dest string) error {
	r, err := zip.NewReader(data, size)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return err
		}
		out.Close()
		rc.Close()
	}
	return nil
}
