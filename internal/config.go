// Package internal loads the configuration and contains note management functions.
package internal

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type TNConfig struct {
	StorageMode         string `toml:"storage_mode"`
	MainVault           string `toml:"path_to_main_vault"`
	TemplateNote        string `toml:"path_to_template_note"`
	Editor              string `toml:"editor"`
	Path_to_inline_note string `toml:"path_to_inline_note"`
}

type Note struct {
	FileName string
	FilePath string
}

func Config(pathToVault string) TNConfig {
	if pathToVault == "" {
		pathToVault = "config.toml"
	}
	var config TNConfig
	_, err := toml.DecodeFile(pathToVault, &config)
	if err != nil {
		fmt.Printf("Config not found from path: %s\n", pathToVault)
		panic(err)
	}
	return config
}

var config TNConfig

func init() {
	config = Config("")
}
