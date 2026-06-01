// Package cmd contains implementation of console commands for performing actions with notes
/*
Copyright © 2026 Sergey <belousovsergej56@gmail.com>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tn/internal"
)

var rootCmd = &cobra.Command{
	Use:   "tn",
	Short: "A fast CLI note manager featuring search, Git, and Obsidian",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "config" || cmd.Name() == "sync" || cmd.Name() == "git-sync" || cmd.Name() == "help" {
			return
		}

		vaultDir := internal.GlobalConfig.MainVault
		if _, err := os.Stat(filepath.Join(vaultDir, ".git")); err != nil {
			return
		}

		_ = exec.Command("git", "-C", vaultDir, "config", "local", "core.filemode", "false").Run()
		_ = exec.Command("git", "-C", vaultDir, "config", "local", "core.autocrlf", "true").Run()

		if !internal.HasRemote(vaultDir) {
			return
		}
		if cmd.Name() == "sync" {
			return
		}

		pullCmd := exec.Command("git", "-C", vaultDir, "pull", "--rebase")
		if err := pullCmd.Run(); err != nil {
			_ = exec.Command("git", "-C", vaultDir, "rebase", "--abort").Run()
			fmt.Println("\x1b[33m[Git Warning] You have divergent branches and need to specify how to reconcile them. It is recommended to check the status manually.\x1b[0m")
		}
	},

	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		mutatingCommands := map[string]bool{
			"inline": true, "edit": true, "grep": true, "new": true, "delete": true,
		}

		if !mutatingCommands[cmd.Name()] {
			return
		}

		vaultDir := internal.GlobalConfig.MainVault
		if _, err := os.Stat(filepath.Join(vaultDir, ".git")); err != nil {
			return
		}

		bgCmd := exec.Command(os.Args[0], "git-sync")
		bgCmd.Stdout = nil
		bgCmd.Stderr = nil
		bgCmd.Stdin = nil

		_ = bgCmd.Start()
	},
}

var gitSyncCmd = &cobra.Command{
	Use:    "git-sync",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		vaultDir := internal.GlobalConfig.MainVault

		_ = exec.Command("git", "-C", vaultDir, "add", ".").Run()

		timestamp := time.Now().Format("2006-01-02 15:04:05")
		commitMsg := fmt.Sprintf("tn: auto-sync %s", timestamp)

		_ = exec.Command("git", "-C", vaultDir, "commit", "-m", commitMsg).Run()

		if internal.HasRemote(vaultDir) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			pushCmd := exec.CommandContext(ctx, "git", "-C", vaultDir, "push")
			_ = pushCmd.Run()
		}
	},
}

var inlineCmd = &cobra.Command{
	Use:     "inline [text]",
	Aliases: []string{"i"},
	Short:   "Create a quick timestamped note directly from arguments",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		noteText := strings.Join(args, " ")
		internal.ValidateVault()
		internal.InlineNote(noteText)
	},
}

var editCmd = &cobra.Command{
	Use:     "edit",
	Aliases: []string{"e"},
	Short:   "Interactive search and edit a note in the terminal",
	Run: func(cmd *cobra.Command, args []string) {
		internal.ValidateVault()
		file := internal.GetFile()
		internal.EditFile(file)
	},
}

var grepCmd = &cobra.Command{
	Use:     "grep",
	Aliases: []string{"g"},
	Short:   "Interactive full-text search across notes",
	Long:    "Launch an interactive fuzzy search (fzf) through the content of all your notes and open the selected file in your editor",
	Run: func(cmd *cobra.Command, args []string) {
		internal.ValidateVault()
		internal.InteractiveSearchInternal()
	},
}

var newCmd = &cobra.Command{
	Use:     "new [filename]",
	Aliases: []string{"n"},
	Args:    cobra.ExactArgs(1),
	Short:   "Create a new note and open it in the terminal",
	Long: `Create a new note and open it in the terminal. 

If the file already exists, it will be opened directly. If it doesn't, 
a new file will be created. If a template file is specified in your 
configuration, its content will be automatically pre-filled into the new note.`,
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		if filepath.Ext(filename) == "" {
			defaultExt := internal.Config("").FileExtension
			if !strings.HasPrefix(defaultExt, ".") {
				defaultExt = "." + defaultExt
			}
			filename = filename + defaultExt
		}
		internal.ValidateVault()
		internal.HandlerFile(filename)
	},
}

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	Short:   "Interactive search and delete a note",
	Run: func(cmd *cobra.Command, args []string) {
		internal.ValidateVault()
		file := internal.GetFile()
		if file == "" {
			fmt.Println("Operation canceled")
			return
		}
		err := internal.RemoveFile(file)
		if err != nil {
			fmt.Printf("\x1b[31m Failed to delete file %s\x1b[0m, error: %s\n", filepath.Base(file), err)
			return
		}
		fmt.Printf("\033[32mFile %s deleted successfully\033[0m\n", filepath.Base(file))
	},
}

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"c"},
	Short:   "Open config file for edit",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, err := internal.GetConfigPath()
		if err != nil {
			fmt.Println(err)
			return
		}
		internal.EditFile(configFile)
	},
}

var openObsidianCmd = &cobra.Command{
	Use:     "open",
	Aliases: []string{"o"},
	Short:   "Interactive search and open a note in Obsidian",
	Run: func(cmd *cobra.Command, args []string) {
		internal.ValidateVault()
		file := internal.GetFile()
		internal.OpenURI(file)
	},
}

var syncCmd = &cobra.Command{
	Use:     "sync",
	Aliases: []string{"s"},
	Short:   "Manually synchronize your notes vault with the remote Git repository",
	Run: func(cmd *cobra.Command, args []string) {
		err := internal.SyncGit()
		if err != nil {
			return
		}
	},
}

func Execute() {
	if len(os.Args) > 1 {
		firstArgs := os.Args[1]
		if !isKnownCommand(firstArgs) && !strings.HasPrefix(firstArgs, "-") {
			os.Args = append([]string{os.Args[0], "inline"}, os.Args[1:]...)
		}
	}
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func isKnownCommand(arg string) bool {
	known := map[string]bool{
		"open": true, "o": true,
		"new": true, "n": true,
		"grep": true, "g": true,
		"inline": true, "i": true,
		"delete": true, "d": true,
		"config": true, "c": true,
		"edit": true, "e": true,
		"sync": true, "s": true,
		"git-sync": true,
		"help":     true,
	}
	return known[arg]
}

func init() {
	rootCmd.AddCommand(
		inlineCmd,
		editCmd,
		grepCmd,
		newCmd,
		deleteCmd,
		configCmd,
		openObsidianCmd,
		syncCmd,
		gitSyncCmd,
	)

}
