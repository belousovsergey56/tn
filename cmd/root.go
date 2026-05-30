// Package cmd contains implementation of console commands for performing actions with notes
/*
Copyright © 2026 Sergey <belousovsergej56@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tn/internal"
)

var rootCmd = &cobra.Command{
	Use:   "tn",
	Short: "A fast CLI note manager featuring search, Git, and Obsidian",
}

var inlineCmd = &cobra.Command{
	Use:     "inline [text]",
	Aliases: []string{"i"},
	Short:   "Create a quick note from text without opening an editor",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		noteText := strings.Join(args, " ")
		internal.ValidateVault()
		internal.InlineNote(noteText)
	},
}

var openCmd = &cobra.Command{
	Use:     "open",
	Aliases: []string{"o"},
	Short:   "Open and edit",
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
	Short:   "Create a new note or open an existing one",
	Long: `Create a new note and open it for editing in your terminal editor. 

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

var removeCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"r"},
	Short:   "Search and deleted",
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
		"remove": true, "r": true,
		"config": true, "c": true,
		"help": true,
	}
	return known[arg]
}

func init() {
	rootCmd.AddCommand(
		inlineCmd,
		openCmd,
		grepCmd,
		newCmd,
		removeCmd,
		configCmd,
	)

}
