package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func HasRemote(vaultDir string) bool {
	out, err := exec.Command("git", "-C", vaultDir, "remote").Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

func SyncGit() error {
	vaultDir := GlobalConfig.MainVault

	if _, err := os.Stat(filepath.Join(vaultDir, ".git")); err != nil {
		fmt.Println("\x1b[31m[Error] The Git repository was not found in the notes folder.\x1b[0m")
		return err
	}

	fmt.Println("Run manual storage synchronization...")

	fmt.Print("Pull changes... ")
	pull := exec.Command("git", "-C", vaultDir, "pull", "--rebase")
	if err := pull.Run(); err != nil {
		_ = exec.Command("git", "-C", vaultDir, "rebase", "--abort").Run()
		fmt.Println("\n\x1b[31mPull error! There may be conflicts. Please check the repository manually.\x1b[0m")
		return err
	}
	fmt.Println("\033[32mSuccessfully\033[0m")

	fmt.Print("Commit local changes... ")
	_ = exec.Command("git", "-C", vaultDir, "add", ".").Run()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	commitMsg := fmt.Sprintf("tn: manual-sync %s", timestamp)
	_ = exec.Command("git", "-C", vaultDir, "commit", "-m", commitMsg).Run()
	fmt.Println("\033[32mSuccessfully\033[0m")

	remoteCheck, err := exec.Command("git", "-C", vaultDir, "remote").Output()
	if err == nil && len(strings.TrimSpace(string(remoteCheck))) > 0 {
		fmt.Print("Sending changes to the server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		push := exec.CommandContext(ctx, "git", "-C", vaultDir, "push")
		if err := push.Run(); err != nil {
			fmt.Println("\n\x1b[31m Push error! Check your internet connection.\x1b[0m")
			return err
		}
		fmt.Println("\033[32mSuccessfully\033[0m")
	} else {
		fmt.Println("The remote repository is not configured. Skip Push.")
	}

	fmt.Println("\x1b[32mSynchronization completed successfully. Storage is up to date\x1b[0m")
	return nil
}
