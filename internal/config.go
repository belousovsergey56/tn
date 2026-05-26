// Package internal loads the configuration and contains note management functions.
package internal

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type TNConfig struct {
	StorageMode  string `toml:"storage_mode"`
	MainValut    string `toml:"path_to_main_valut"`
	TemplateNote string `toml:"path_to_template_note"`
	Editor       string `toml:"editor"`
}

type Note struct {
	FileName string
	FilePath string
}

func Config() TNConfig {
	var config TNConfig
	_, err := toml.DecodeFile("./config.toml", &config)
	if err != nil {
		fmt.Println("Config not found from path $HOME/.config/tn/config.toml")
		panic(err)
	}
	return config
}
