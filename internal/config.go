// Package internal loads the configuration and contains note management functions.
package internal

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

//go:embed default_config.toml
var defaultConfigFile []byte

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	baseConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseConfigDir, "terminal-note", "config.toml"), nil
}

// EnsureConfigExists checks whether the directory and config file exist and creates them
func EnsureConfigExists() error {
	baseConfigDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	appDir := filepath.Join(baseConfigDir, "terminal-note")
	configPath := filepath.Join(appDir, "config.toml")

	err = os.MkdirAll(appDir, 0755)
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err = os.WriteFile(configPath, defaultConfigFile, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// TNConfig contains fields for working with config file data
type TNConfig struct {
	StorageMode      string `toml:"storage_mode"`
	MainVault        string `toml:"path_to_main_vault"`
	TemplateNote     string `toml:"path_to_template_note"`
	Editor           string `toml:"editor"`
	PathToInlineNote string `toml:"path_to_inline_note"`
	FileExtension    string `toml:"file_extension"`
}

// Note used to work with files
type Note struct {
	FileName string
	FilePath string
}

// SearchLine used to interactively search text in files
type SearchLine struct {
	FilePath string
	Line     string
	Col      string
	Text     string
}

// Config parses config.toml and returns TNConfig struct
func Config(pathToVault string) TNConfig {
	if pathToVault == "" {
		sysPath, err := GetConfigPath()
		if err != nil {
			panic(err)
		}
		pathToVault = sysPath
	}
	var config TNConfig
	_, err := toml.DecodeFile(pathToVault, &config)
	if err != nil {
		fmt.Printf("Config not found from path: %s\n", pathToVault)
		panic(err)
	}
	config.MainVault = ExpandPath(config.MainVault)
	config.PathToInlineNote = ExpandPath(config.PathToInlineNote)
	config.TemplateNote = ExpandPath(config.TemplateNote)
	return config
}

// ExpandPath expands ~ and environment variables like $HOME into a full path,
func ExpandPath(path string) string {
	if path == "" {
		return ""
	}

	path = os.ExpandEnv(path)
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}

	return path
}

// ValidateVault checks if the vault exists and handles its absence
func ValidateVault() {
	vaultPath := GlobalConfig.MainVault
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		isDefault := vaultPath == ExpandPath("~/terminal-note")

		if isDefault {
			err := os.MkdirAll(vaultPath, 0755)
			if err != nil {
				fmt.Printf("\x1b[31mFailed to create default storage: %v\x1b[0m\n", err)
				os.Exit(1)
			}
			if GlobalConfig.PathToInlineNote != "" {
				os.MkdirAll(GlobalConfig.PathToInlineNote, 0755)
			}
			fmt.Printf("Hi, a new note repository is automatically created: %s\n", vaultPath)
		} else {
			fmt.Printf("\x1b[31mError: The specified storage folder does not exist:\x1b[0m %s\n", vaultPath)
			fmt.Println("Please create it manually or correct the path in the config: `tn config`")
			os.Exit(1)
		}
	}
}

var GlobalConfig TNConfig

func init() {
	err := EnsureConfigExists()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize config folder: %v", err))
	}
	GlobalConfig = Config("")
}
