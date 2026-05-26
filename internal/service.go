package internal

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
)

func ScanValute(root string) []Note {
	var notes []Note
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "md") {
				note := Note{FileName: d.Name(), FilePath: path}
				notes = append(notes, note)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("scan error %s\n", err)
	}
	return notes
}

func EditFile(filePath string) {
	cmd := exec.Command("hx", filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("command ended with error, %v\n", err)
		return
	}
}

func GetFile() string {
	noteList := ScanValute(Config().MainValut)
	idx, err := fuzzyfinder.Find(
		noteList,
		func(i int) string {
			return noteList[i].FileName
		},

		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			contentBytes, err := os.ReadFile(noteList[i].FilePath)
			if err != nil {
				return fmt.Sprintf("Ошибка чтения файла: %v", err)
			}
			fileContent := string(contentBytes)
			return fmt.Sprintf("=== %s ===\nПуть: %s\n\n%s",
				noteList[i].FileName,
				noteList[i].FilePath,
				fileContent,
			)
		}))
	if err != nil {
		log.Fatal(err)
	}
	return noteList[idx].FilePath
}
