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

func getRelativePathSafe(fullPath, vaultName string) string {
	cleanPath := filepath.Clean(fullPath)
	lookFor := string(filepath.Separator) + vaultName + string(filepath.Separator)
	index := strings.Index(cleanPath, lookFor)

	if index == -1 {
		if strings.HasPrefix(cleanPath, vaultName+string(filepath.Separator)) {
			return strings.TrimPrefix(cleanPath, vaultName+string(filepath.Separator))
		}
		if strings.HasSuffix(cleanPath, vaultName) {
			return ""
		}
		return cleanPath
	}

	relative := cleanPath[index+len(lookFor):]
	relative = strings.TrimSuffix(relative, filepath.Ext(relative))
	return filepath.ToSlash(relative)
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

func OpenURI(vaultName, notePath string) error {
	escapedVault := strictEscape(vaultName)
	var escapedFile string
	var obsidianURI string
	var cmd *exec.Cmd

	if isWsl() {
		relativePath := getRelativePathSafe(notePath, vaultName)
		escapedFile = strictEscape(relativePath)
		obsidianURI = fmt.Sprintf("obsidian://open?vault=%s&file=%s", escapedVault, escapedFile)
		fmt.Println(escapedVault)
		fmt.Println(escapedFile)
		cmd = exec.Command("powershell.exe", "-NoProfile", "-Command", "Start-Process",
			fmt.Sprintf("'%s'", obsidianURI))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	escapedFile = url.QueryEscape(notePath)
	obsidianURI = fmt.Sprintf("obsidian://open?vault=%s&file=%s", escapedVault, escapedFile)
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", obsidianURI)
	case "darwin":
		cmd = exec.Command("open", obsidianURI)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Run()

}
