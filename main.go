/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	// "tn/cmd"
	"fmt"
	"path/filepath"

	// "runtime"
	"tn/internal"
)

func main() {
	// fmt.Println(runtime.GOOS)
	file := internal.GetFile()
	// internal.EditFile(file)
	valuteName := filepath.Base(internal.Config().MainValut)
	// fmt.Println(valuteName)
	err := internal.OpenURI(valuteName, file)
	if err != nil {
		fmt.Println("Ошибка при открытии Obsidian:", err)
	}
	// cmd.Execute()

}
