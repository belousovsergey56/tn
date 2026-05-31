package internal

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func getRelativePathSafe(notePath, vaultPath string) (string, error) {
	relative, err := filepath.Rel(vaultPath, notePath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path %s", err)
	}

	if strings.HasPrefix(relative, "..") {
		return "", fmt.Errorf("the file is outside vault")
	}

	relative = strings.TrimSuffix(relative, filepath.Ext(relative))
	return filepath.ToSlash(relative), nil
}

func isWsl() bool {
	os, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(os)), "microsoft")
}

func strictEscape(s string) string {
	s = url.QueryEscape(s)
	s = strings.ReplaceAll(s, "+", "%20")
	return s
}

func OpenURI(notePath string) error {
	vaultPath := GlobalConfig.MainVault
	if vaultPath == "" {
		return fmt.Errorf("MainVault is not configured")
	}
	vaultName := filepath.Base(vaultPath)
	relativePath, err := getRelativePathSafe(notePath, vaultPath)
	if err != nil {
		return err
	}

	escapedVault := strictEscape(vaultName)
	escapedFile := strictEscape(relativePath)

	obsidianURI := fmt.Sprintf("obsidian://open?vault=%s&file=%s", escapedVault, escapedFile)

	if isWsl() {
		return openInWsl(obsidianURI)
	}

	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", obsidianURI).Run()
	case "darwin":
		return exec.Command("open", obsidianURI).Run()
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
}

func openInWsl(uri string) error {
	return exec.Command("powershell.exe", "-NoProfile", "-Command", "Start-Process",
		fmt.Sprintf("'%s'", uri)).Run()
}
