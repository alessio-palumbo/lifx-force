package runtime

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "embed"
)

//go:embed assets/fingertrack.zip
var fingertrackZip []byte

// EnsureFingertrackInstalled makes sure the correct version is extracted.
func EnsureFingertrackInstalled() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	baseDir := filepath.Join(home, ".lifx-force")
	ftPath := filepath.Join(baseDir, "fingertrack")
	versionFile := filepath.Join(baseDir, ".version")

	// Check installed version matches current otherwise replace.
	if data, err := os.ReadFile(versionFile); err == nil {
		if strings.TrimSpace(string(data)) == FingertrackVersion {
			return ftPath, nil
		}
	}

	if err := unzip(bytes.NewReader(fingertrackZip), int64(len(fingertrackZip)), baseDir); err != nil {
		return "", err
	}

	if err := os.WriteFile(versionFile, []byte(FingertrackVersion), 0644); err != nil {
		return "", err
	}

	return ftPath, nil
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
